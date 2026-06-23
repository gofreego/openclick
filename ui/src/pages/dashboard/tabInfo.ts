import type { TabInfo } from '../../components/TabInfoButton'

export const COLORS = ['#6366f1', '#f59e0b', '#10b981', '#ef4444', '#8b5cf6', '#06b6d4']

export const TAB_INFO: Record<string, TabInfo> = {
  trends: {
    title: 'Trends Query',
    meaning: 'Trends let you view and analyze event counts, frequencies, and metrics over time. It answers questions like "How many times was page X viewed?" or "What is our daily active user count?"',
    howToUse: 'Enter the event name you want to track (e.g. $pageview), select the date range and interval (Hour, Day, Week, Month), and click "Run Query". You can configure customized events as needed.',
    example: 'To check signup conversions: Run a query with event "user_signed_up" over the "Last 30 Days" with a "Day" interval to see a day-by-day graph of completed registration events.'
  },
  funnel: {
    title: 'Funnel Query',
    meaning: 'Funnels measure how users complete a series of defined steps in your app, helping you visualize conversion and drop-off rates at each stage of a specific user flow.',
    howToUse: 'Add steps in chronological order. Specify the event key and a user-friendly label for each step. Define the date range and a "Conversion Window" (the max time a user has to complete all steps to count as converted), then click "Run Funnel".',
    example: 'For a purchase checkout funnel, set: Step 1 (cart_viewed) -> Step 2 (checkout_started) -> Step 3 (payment_submitted) with a 14-day conversion window to identify where users drop off.'
  },
  retention: {
    title: 'Retention Query',
    meaning: 'Retention metrics track user engagement over time. It measures how many users who completed a starting action return to perform another key action in subsequent days, weeks, or months.',
    howToUse: 'Select the "Target Event" (the starting event that places users into a cohort) and the "Return Event" (the activity that marks them as retained). Choose the date range and the grouping period (Day, Week, Month), then click "Run Retention".',
    example: 'To check weekly user return rate: Set Target Event to "user_signed_up" and Return Event to "$pageview" with a "Week" period. The cohort matrix will show the percentage of users returning in Week 1, Week 2, Week 3, etc.'
  },
  paths: {
    title: 'User Paths Query',
    meaning: 'User Paths visualize the step-by-step journeys and flows users follow through your website or application, revealing the most common routes taken between screens or events.',
    howToUse: 'Choose a date range and specify the limit of steps. You can optionally filter paths starting at a specific page (Start Point) or ending at a specific page (End Point). Click "Run Paths" to view the resulting flows and nodes.',
    example: 'To inspect paths after pricing page: Set Start Point to "/pricing" and Step Limit to 5. The path flow results will illustrate the next 5 pages or events users visited immediately after viewing the pricing details.'
  },
  dashboards: {
    title: 'Dashboards',
    meaning: 'Dashboards allow you to combine, organize, and monitor saved charts and metrics widgets in a single, consolidated page.',
    howToUse: 'Click "Create Dashboard", input a name, and open it to view saved reports. You can save queries from other pages directly to your custom dashboards to keep key metrics visible.',
    example: 'Create a "Key KPIs" dashboard and pin widgets like "Daily Active Users (Trends)", "Sign-up to Paid Funnel", and "Feature Flag active rollouts" to monitor product performance at a glance.'
  }
}
