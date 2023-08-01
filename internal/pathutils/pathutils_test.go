package pathutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindCommonParent(t *testing.T) {
	tests := []struct {
		name       string
		filePaths  []string
		commonPath string
	}{
		{
			name: "Multiple Files",
			filePaths: []string{
				"layers/eks.tf",
				"layers/eks/main.tf",
				"layers/eks/output.tf",
			},
			commonPath: "layers",
		},
		{
			name: "Single File",
			filePaths: []string{
				"layers/eks/output.tf",
			},
			commonPath: "layers/eks",
		},
		{
			name: "No Common Path",
			filePaths: []string{
				"file1.tf",
				"file2.tf",
			},
			commonPath: ".",
		},
		{
			name: "Different Levels",
			filePaths: []string{
				"layers/eks.tf",
				"modules/eks/main.tf",
			},
			commonPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commonParent := FindCommonParentPath(tt.filePaths)
			assert.Equal(t, tt.commonPath, commonParent)
		})
	}
}
