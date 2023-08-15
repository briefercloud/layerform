package tags

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

func AddTagsToFile(filePath string, tagsToAdd map[string]string) error {
	mapVal := map[string]cty.Value{}
	for k, v := range tagsToAdd {
		mapVal[k] = cty.StringVal(v)
	}
	tags := cty.MapVal(mapVal)

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	hclFile, diags := hclwrite.ParseConfig(fileBytes, filePath, hcl.InitialPos)
	if diags.HasErrors() {
		return diags
	}

	modified := false
	blocks := hclFile.Body().Blocks()

	for _, block := range blocks {
		if block.Type() == "resource" && strings.HasPrefix(block.Labels()[0], "aws_") {
			modified = true

			tagsAttr := block.Body().GetAttribute("tags")
			if tagsAttr == nil {
				_ = block.Body().SetAttributeValue("tags", tags)
				continue
			}

			expr := tagsAttr.Expr()
			tokens := expr.BuildTokens(nil)

			merge := hclwrite.TokensForFunctionCall(
				"merge",
				hclwrite.TokensForValue(tags),
				tokens,
			)
			_ = block.Body().SetAttributeRaw("tags", merge)
		}
	}

	if modified {
		err = os.WriteFile(filePath, hclFile.Bytes(), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
