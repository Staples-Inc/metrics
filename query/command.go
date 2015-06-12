// Copyright 2015 Square Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package query

import (
	"sort"

	"github.com/square/metrics/api"
)

// Command is the final result of the parsing.
// A command contains all the information to execute the
// given query against the API.
type Command interface {
	// Execute the given command. Returns JSON-encodable result or an error.
	Execute(api.MultiBackend, api.API) (interface{}, error)
	Name() string
}

// DescribeCommand describes the tag set managed by the given metric indexer.
type DescribeCommand struct {
	metricName api.MetricKey
	predicate  api.Predicate
}

// DescribeAllCommand returns all the metrics available in the system.
type DescribeAllCommand struct {
}

// SelectCommand is the bread and butter of the metrics query engine.
// It actually performs the query against the underlying metrics system.
type SelectCommand struct {
	predicate   api.Predicate
	expressions []Expression
	context     *evaluationContextNode
}

// Execute returns the list of tags satisfying the provided predicate.
func (cmd *DescribeCommand) Execute(b api.MultiBackend, a api.API) (interface{}, error) {
	tags, _ := a.GetAllTags(cmd.metricName)
	output := make([]string, 0, len(tags))
	for _, tag := range tags {
		if cmd.predicate.Apply(tag) {
			output = append(output, tag.Serialize())
		}
	}
	sort.Strings(output)
	return output, nil
}
func (cmd *DescribeCommand) Name() string {
	return "describe"
}

// Execute of a DescribeAllCommand returns the list of all metrics.
func (cmd *DescribeAllCommand) Execute(b api.MultiBackend, a api.API) (interface{}, error) {
	result, err := a.GetAllMetrics()
	if err == nil {
		sort.Sort(api.MetricKeys(result))
	}
	return result, err
}
func (cmd *DescribeAllCommand) Name() string {
	return "describe all"
}

// Execute performs the query represented by the given query string, and returs the result.
func (cmd *SelectCommand) Execute(b api.MultiBackend, a api.API) (interface{}, error) {
	timerange, err := api.NewSnappedTimerange(cmd.context.Start, cmd.context.End, cmd.context.Resolution)
	if err != nil {
		return nil, err
	}
	return evaluateExpressions(EvaluationContext{
		MultiBackend: b,
		Timerange:    timerange,
		SampleMethod: cmd.context.SampleMethod,
		Predicate:    cmd.predicate,
		API:          a,
	}, cmd.expressions)
}
func (cmd *SelectCommand) Name() string {
	return "select"
}
