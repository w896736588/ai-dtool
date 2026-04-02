package plw

import "testing"

// TestParseLocatorInputSpec 验证结构化 Locator 配置会被归一化成标准格式。
func TestParseLocatorInputSpec(t *testing.T) {
	parser := NewLocatorParser()
	exact := true
	first := true

	input := &LocatorInput{
		Spec: &LocatorSpec{
			FindType: `role`,
			Value:    `button`,
			Name:     `提交`,
			Exact:    &exact,
			First:    &first,
			Chain: []LocatorSpec{
				{
					Method: `locator`,
					Value:  `.icon-save`,
				},
			},
		},
	}

	spec, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if spec.Method != `role` {
		t.Fatalf("spec.Method = %q, want %q", spec.Method, `role`)
	}
	if spec.Options == nil || spec.Options.Name != `提交` {
		t.Fatalf("spec.Options.Name = %q, want %q", spec.Options.Name, `提交`)
	}
	if spec.Options.Exact == nil || !*spec.Options.Exact {
		t.Fatalf("spec.Options.Exact = %#v, want true", spec.Options.Exact)
	}
	if spec.Pick == nil || spec.Pick.First == nil || !*spec.Pick.First {
		t.Fatalf("spec.Pick.First = %#v, want true", spec.Pick)
	}
	if len(spec.Chain) != 1 || spec.Chain[0].Method != `locator` {
		t.Fatalf("spec.Chain = %#v, want one locator chain", spec.Chain)
	}
}

// TestParseLocatorInputPickConflict 验证 first、last、nth 互斥。
func TestParseLocatorInputPickConflict(t *testing.T) {
	parser := NewLocatorParser()
	first := true
	last := true

	_, err := parser.Parse(&LocatorInput{
		Spec: &LocatorSpec{
			Method: `text`,
			Value:  `立即提交`,
			First:  &first,
			Last:   &last,
		},
	})
	if err == nil {
		t.Fatal("Parse() error = nil, want pick conflict error")
	}
}

// TestParseLocatorInputMissingMethod 验证缺少 method 时会报错。
func TestParseLocatorInputMissingMethod(t *testing.T) {
	parser := NewLocatorParser()

	_, err := parser.Parse(&LocatorInput{
		Spec: &LocatorSpec{
			Value: `提交`,
		},
	})
	if err == nil {
		t.Fatal("Parse() error = nil, want missing method error")
	}
}

// TestParseLocatorInputRequireStructured 验证 LocatorParser 只接受结构化配置。
func TestParseLocatorInputRequireStructured(t *testing.T) {
	parser := NewLocatorParser()

	_, err := parser.Parse(&LocatorInput{})
	if err == nil {
		t.Fatal("Parse() error = nil, want structured locator required error")
	}
}
