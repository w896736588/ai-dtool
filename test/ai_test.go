package test

import (
	"gitee.com/Sxiaobai/gs/gstool"
	"strings"
	"testing"
)

// TestBailian 百炼 qwen2.5-coder-3b-instruct 模型
func TestBailian(t *testing.T) {
	s := `remote: Compressing objects:   0% (1/236)[K
remote: Compressing objects:   1% (3/236)[K
remote: Compressing objects:   2% (5/236)[K
remote: Compressing objects:   3% (8/236)[K
remote: Compressing objects:   4% (10/236)[K
remote: Compressing objects:   5% (12/236)[K
remote: Compressing objects:   6% (15/236)[K
remote: Compressing objects:   7% (17/236)[K
remote: Compressing objects:   8% (19/236)[K
remote: Compressing objects:   9% (22/236)[K
remote: Compressing objects:  10% (24/236)[K
remote: Compressing objects:  11% (26/236)[K
remote: Compressing objects:  12% (29/236)[K
remote: Compressing objects:  13% (31/236)[K
remote: Compressing objects:  14% (34/236)[K
remote: Compressing objects:  15% (36/236)[K
remote: Compressing objects:  16% (38/236)[K
remote: Compressing objects:  17% (41/236)[K
remote: Compressing objects:  18% (43/236)[K
remote: Compressing objects:  19% (45/236)[K
remote: Compressing objects:  20% (48/236)[K
remote: Compressing objects:  21% (50/236)[K
remote: Compressing objects:  22% (52/236)[K
remote: Compressing objects:  23% (55/236)[K
remote: Compressing objects:  24% (57/236)[K
remote: Compressing objects:  25% (59/236)[K
remote: Compressing objects:  26% (62/236)[K
remote: Compressing objects:  27% (64/236)[K
remote: Compressing objects:  28% (67/236)[K
remote: Compressing objects:  29% (69/236)[K
remote: Compressing objects:  30% (71/236)[K
remote: Compressing objects:  31% (74/236)[K
remote: Compressing objects:  32% (76/236)[K
remote: Compressing objects:  33% (78/236)[K
remote: Compressing objects:  34% (81/236)[K
remote: Compressing objects:  35% (83/236)[K
remote: Compressing objects:  36% (85/236)[K
remote: Compressing objects:  37% (88/236)[K
remote: Compressing objects:  38% (90/236)[K
remote: Compressing objects:  39% (93/236)[K
remote: Compressing objects:  40% (95/236)[K
remote: Compressing objects:  41% (97/236)[K
remote: Compressing objects:  42% (100/236)[K
remote: Compressing objects:  43% (102/236)[K
remote: Compressing objects:  44% (104/236)[K
remote: Compressing objects:  45% (107/236)[K
remote: Compressing objects:  46% (109/236)[K
remote: Compressing objects:  47% (111/236)[K
remote: Compressing objects:  48% (114/236)[K
remote: Compressing objects:  49% (116/236)[K
remote: Compressing objects:  50% (118/236)[K
remote: Compressing objects:  51% (121/236)[K
remote: Compressing objects:  52% (123/236)[K
remote: Compressing objects:  53% (126/236)[K
remote: Compressing objects:  54% (128/236)[K
remote: Compressing objects:  55% (130/236)[K
remote: Compressing objects:  56% (133/236)[K
remote: Compressing objects:  57% (135/236)[K
remote: Compressing objects:  58% (137/236)[K
remote: Compressing objects:  59% (140/236)[K
remote: Compressing objects:  60% (142/236)[K
remote: Compressing objects:  61% (144/236)[K
remote: Compressing objects:  62% (147/236)[K
remote: Compressing objects:  63% (149/236)[K
remote: Compressing objects:  64% (152/236)[K
remote: Compressing objects:  65% (154/236)[K
remote: Compressing objects:  66% (156/236)[K
remote: Compressing objects:  67% (159/236)[K
remote: Compressing objects:  68% (161/236)[K
remote: Compressing objects:  69% (163/236)[K
remote: Compressing objects:  70% (166/236)[K
remote: Compressing objects:  71% (168/236)[K`
	gstool.FmtPrintlnLogTime(`%s`, gstool.JsonEncode(strings.Split(s, `[K`)))
}
