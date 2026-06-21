# OpenClick 

## Overview

**OpenClick** is an open-source, self-hostable product analytics platform built for speed and low memory footprint. It provides event tracking, session replay, funnel analysis, feature flags, and cohort analytics.

OpenClick is written in **Go** for the backend (single binary, low GC pressure, high concurrency) and **React + TypeScript** for the dashboard UI. The storage layer uses **ClickHouse** for analytics queries and **PostgreSQL** for relational metadata.
