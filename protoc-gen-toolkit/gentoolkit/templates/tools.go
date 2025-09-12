/*
Copyright [2014] - [2023] The Last.Backend authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package templates

// ToolsContentTpl generates tools initialization template
var ToolsContentTpl = `
	// Start metrics server with default provider
	if err := app.runtime.Tools().StartMetrics(context.Background(), "prometheus"); err != nil {
		return nil, fmt.Errorf("failed to start metrics: %w", err)
	}

	// Start probes server (always enabled by default)  
	if err := app.runtime.Tools().StartProbes(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to start probes: %w", err)
	}
`
