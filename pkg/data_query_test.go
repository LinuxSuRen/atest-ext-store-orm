/*
Copyright 2025 API Testing Authors.

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
package pkg

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunMultilineSQL(t *testing.T) {
	tests := []struct {
		name string
		sql  string
	}{
		{
			name: "empty",
			sql:  "",
		}, {
			name: "multiple blanks",
			sql:  "  ; ",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := runMultilineSQL(context.TODO(), test.sql, nil)
			assert.NoError(t, err)
			assert.Nil(t, result)
		})
	}
}
