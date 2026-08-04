// Copyright 2023 Ubie, inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestValidateStringOneOf(t *testing.T) {
	v := ValidateStringOneOf{Allowed: []string{"member", "viewer", "editor"}}

	tests := []struct {
		name     string
		input    types.String
		wantErrs int
	}{
		{name: "allowed member", input: types.StringValue("member"), wantErrs: 0},
		{name: "allowed viewer", input: types.StringValue("viewer"), wantErrs: 0},
		{name: "rejected superadmin", input: types.StringValue("superadmin"), wantErrs: 1},
		{name: "rejected uppercase Member", input: types.StringValue("Member"), wantErrs: 1},
		{name: "null value is skipped", input: types.StringNull(), wantErrs: 0},
		{name: "unknown value is skipped", input: types.StringUnknown(), wantErrs: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validator.StringRequest{ConfigValue: tt.input}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)
			if got := resp.Diagnostics.ErrorsCount(); got != tt.wantErrs {
				t.Errorf("errors = %d, want %d (details: %v)", got, tt.wantErrs, resp.Diagnostics.Errors())
			}
		})
	}
}
