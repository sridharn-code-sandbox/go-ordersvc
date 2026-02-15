// Copyright 2026 go-ordersvc Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package main is the entry point for the ordersvc application.
package main

import (
	"fmt"
	"os"

	"github.com/nsridhar76/go-ordersvc/internal/config"
)

var version = "dev"

func main() {
	// Load configuration
	cfg, err := config.LoadFromEnv()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Set version
	cfg.App.Version = version

	// Run server
	if err := Run(cfg); err != nil {
		fmt.Printf("Server failed: %v\n", err)
		os.Exit(1)
	}
}
