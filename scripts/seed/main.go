package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gofreego/goutils/logger"
)

const apiBase = "http://127.0.0.1/api/v1"

func request(method, endpoint string, body interface{}) (map[string]interface{}, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, apiBase+endpoint, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-user-id", "1")
	req.Header.Set("x-user-perms", "projects:read, projects:write, projects:delete, dashboards:read, dashboards:write, dashboards:delete, members:write, analytics:read, events:read, replay:read, replay:delete, persons:read, persons:delete, flags:read, flags:write, flags:delete")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API Error [%s %s]: %d %s", method, endpoint, resp.StatusCode, string(respBytes))
	}

	var data map[string]interface{}
	if len(respBytes) > 0 {
		_ = json.Unmarshal(respBytes, &data)
	}
	return data, nil
}

func main() {
	ctx := context.Background()
	logger.Info(ctx, "🚀 Starting OpenClick e2e test seed script (Go)...\n")

	// 1. Create a Test Project
	logger.Info(ctx, "📦 1. Creating a test project...")
	projResp, err := request("POST", "/projects", map[string]string{
		"name":     "E2E Test Project",
		"timezone": "UTC",
	})
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}
	projectId, _ := projResp["id"].(string)
	apiKey, _ := projResp["apiKey"].(string)
	if apiKey == "" {
		apiKey, _ = projResp["api_key"].(string) // fallback just in case
	}
	fmt.Printf("✅ Created Project: %v (%s)\n", projResp["name"], projectId)

	// 2. Inject Events
	fmt.Println("\n📊 2. Injecting test events & persons...")
	eventsToInject := []map[string]interface{}{
		{"event": "$pageview", "distinct_id": "user_1", "properties": map[string]interface{}{"url": "/", "browser": "Chrome", "country": "US"}},
		{"event": "$pageview", "distinct_id": "user_1", "properties": map[string]interface{}{"url": "/pricing", "browser": "Chrome", "country": "US"}},
		{"event": "$click", "distinct_id": "user_1", "properties": map[string]interface{}{"button": "signup", "browser": "Chrome", "country": "US"}},
		{"event": "$pageview", "distinct_id": "user_2", "properties": map[string]interface{}{"url": "/", "browser": "Firefox", "country": "UK"}},
		{"event": "purchase", "distinct_id": "user_2", "properties": map[string]interface{}{"plan": "pro", "amount": 29}},
		{"event": "$pageview", "distinct_id": "user_3", "properties": map[string]interface{}{"url": "/", "browser": "Safari", "country": "CA"}},
		{"event": "$pageview", "distinct_id": "user_4", "properties": map[string]interface{}{"url": "/", "browser": "Chrome", "country": "US"}},
		{"event": "$click", "distinct_id": "user_4", "properties": map[string]interface{}{"button": "signup", "browser": "Chrome", "country": "US"}},
	}

	for _, evt := range eventsToInject {
		evt["apiKey"] = apiKey
		evt["timestamp"] = time.Now().Format(time.RFC3339)
		_, err := request("POST", "/e", evt)
		if err != nil {
			fmt.Printf("❌ Failed to inject event: %v\n", err)
			os.Exit(1)
		}
	}
	fmt.Println("✅ Injected 8 test events across 4 users.")

	// 3. Create a Cohort
	fmt.Println("\n👥 3. Creating a test cohort...")
	cohortResp, err := request("POST", "/projects/"+projectId+"/cohorts", map[string]interface{}{
		"name": "US Chrome Users",
		"filters": map[string]interface{}{
			"country": "US",
			"browser": "Chrome",
		},
	})
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Created Cohort: %v\n", cohortResp["name"])

	// 4. Create a Dashboard & Items
	fmt.Println("\n📈 4. Creating a test dashboard...")
	dashResp, err := request("POST", "/projects/"+projectId+"/dashboards", map[string]interface{}{
		"name": "Main KPIs Dashboard",
	})
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}
	dashboardId, _ := dashResp["id"].(string)
	fmt.Printf("✅ Created Dashboard: %v\n", dashResp["name"])

	fmt.Println("   Adding Trends Chart...")
	_, err = request("POST", "/projects/"+projectId+"/dashboards/"+dashboardId+"/items", map[string]interface{}{
		"name": "Pageviews Trend",
		"type": "trends",
		"query": map[string]interface{}{
			"events":   []map[string]interface{}{{"id": "$pageview", "name": "$pageview", "math": "total"}},
			"interval": "day",
		},
	})
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("   Adding Funnel Chart...")
	_, err = request("POST", "/projects/"+projectId+"/dashboards/"+dashboardId+"/items", map[string]interface{}{
		"name": "Signup Funnel",
		"type": "funnel",
		"query": map[string]interface{}{
			"steps": []map[string]interface{}{{"event": "$pageview"}, {"event": "$click"}},
		},
	})
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Added dashboard items.")

	// 5. Test Analytics Queries directly (Trends, Funnel)
	fmt.Println("\n🔍 5. Verifying Analytics Queries...")
	dateFrom := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	dateTo := time.Now().Format("2006-01-02")

	trendsResp, err := request("POST", "/projects/"+projectId+"/query/trends", map[string]interface{}{
		"events":   []map[string]interface{}{{"id": "$pageview", "name": "$pageview", "math": "total"}},
		"dateFrom": dateFrom,
		"dateTo":   dateTo,
		"interval": "day",
		"filters":  []interface{}{},
	})
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}
	seriesLen := 0
	if results, ok := trendsResp["results"].([]interface{}); ok {
		seriesLen = len(results)
	}
	fmt.Printf("✅ Trends Query OK (Found %d series)\n", seriesLen)

	funnelResp, err := request("POST", "/projects/"+projectId+"/query/funnel", map[string]interface{}{
		"steps": []map[string]interface{}{
			{"event": "$pageview", "name": "View"},
			{"event": "$click", "name": "Click"},
		},
		"dateFrom":             dateFrom,
		"dateTo":               dateTo,
		"conversionWindowDays": 14,
		"filters":              []interface{}{},
	})
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}
	stepsLen := 0
	if results, ok := funnelResp["result"].([]interface{}); ok {
		stepsLen = len(results)
	}
	fmt.Printf("✅ Funnel Query OK (Found %d steps)\n", stepsLen)

	// 6. Test Settings/Permissions
	fmt.Println("\n⚙️ 6. Checking system permissions...")
	permsResp, err := request("GET", "/permissions", nil)
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}
	permsLen := 0
	if perms, ok := permsResp["permissions"].([]interface{}); ok {
		permsLen = len(perms)
	}
	fmt.Printf("✅ Permissions Check OK (Found %d permissions)\n", permsLen)

	fmt.Println("\n🎉 ALL SEEDING AND TESTS PASSED!")
	fmt.Printf("👉 Please go to the UI, select the Project \"%s\", and verify the data.\n", projResp["name"])
}
