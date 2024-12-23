package main

import (
	"sync"
	"time"
	"fmt"
)

type Route struct {
	Destination string    `json:"destination"`
	Gateway     string    `json:"gateway"`
	Interface   string    `json:"interface"`
	Metric      string    `json:"metric"`
	ExpiresAt   time.Time `json:"expires_at"`
}

var (
	routeTable    = make(map[string]Route) // API経由の経路情報
	ripRouteTable = make(map[string]Route) // RIP Updateで受信した経路情報
	mu            sync.Mutex
)

// 経路の追加
func addRoute(route Route) {
	fmt.Printf("add Route:%v\n",route)
	route.ExpiresAt = time.Now().Add(180 * time.Second)
	mu.Lock()
	defer mu.Unlock()
	routeTable[route.Destination] = route
}

func holddownRoute(route Route) {
	mu.Lock()
	defer mu.Unlock()
	u,ok := routeTable[route.Destination]
	if !ok {
		return
	}
	u.Metric = "16" //holddown route
	routeTable[route.Destination] = u
}

// 経路の削除
func removeRoute(route Route) {
	mu.Lock()
	defer mu.Unlock()
	delete(routeTable, route.Destination)
}

// 経路の取得
func getRoutes() []Route {
	mu.Lock()
	defer mu.Unlock()

	routes := []Route{}
	for _, route := range routeTable {
		// poison route
		if route.Metric == "16" {
			delete(routeTable, route.Destination)
			continue
		}
		if time.Now().Before(route.ExpiresAt) {
			routes = append(routes, route)
		}
	}
	return routes
}

// RIP経路の追加
func addRIPRoute(route Route) {
	route.ExpiresAt = time.Now().Add(180 * time.Second)
	fmt.Printf("add RIP Route:%v\n",route)
	mu.Lock()
	defer mu.Unlock()
	ripRouteTable[route.Destination] = route
}

// RIP経路の削除
func removeRIPRoute(route Route) {
	mu.Lock()
	defer mu.Unlock()
	delete(ripRouteTable, route.Destination)
}

// RIP経路の取得
func getRIPRoutes() []Route {
	mu.Lock()
	defer mu.Unlock()

	routes := []Route{}
	for _, route := range ripRouteTable {
		if time.Now().Before(route.ExpiresAt) {
			routes = append(routes, route)
		}
	}
	return routes
}
