package region

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"xray-cli/subscription"
)

var regionKeywords = map[string][]string{
	"香港": {"香港", "HK", "Hong Kong", "🇭🇰"},
	"台湾": {"台湾", "TW", "Taiwan", "🇹🇼"},
	"日本": {"日本", "JP", "Japan", "🇯🇵"},
	"新加坡": {"新加坡", "SG", "Singapore", "🇸🇬"},
	"韩国": {"韩国", "KR", "Korea", "🇰🇷"},
	"美国": {"美国", "US", "USA", "United States", "🇺🇸"},
	"英国": {"英国", "GB", "UK", "United Kingdom", "🇬🇧"},
	"德国": {"德国", "DE", "Germany", "🇩🇪"},
	"法国": {"法国", "FR", "France", "🇫🇷"},
	"澳大利亚": {"澳大利亚", "AU", "Australia", "🇦🇺"},
	"加拿大": {"加拿大", "CA", "Canada", "🇨🇦"},
	"俄罗斯": {"俄罗斯", "RU", "Russia", "🇷🇺"},
	"印度": {"印度", "IN", "India", "🇮🇳"},
	"巴西": {"巴西", "BR", "Brazil", "🇧🇷"},
	"土耳其": {"土耳其", "TR", "Turkey", "🇹🇷"},
	"马来西亚": {"马来西亚", "MY", "Malaysia", "🇲🇾"},
	"泰国": {"泰国", "TH", "Thailand", "🇹🇭"},
	"越南": {"越南", "VN", "Vietnam", "🇻🇳"},
	"菲律宾": {"菲律宾", "PH", "Philippines", "🇵🇭"},
	"印尼": {"印尼", "印度尼西亚", "ID", "Indonesia", "🇮🇩"},
	"意大利": {"意大利", "IT", "Italy", "🇮🇹"},
	"智利": {"智利", "CL", "Chile", "🇨🇱"},
	"尼日利亚": {"尼日利亚", "NG", "Nigeria", "🇳🇬"},
	"澳门": {"澳门", "MO", "Macao", "🇲🇴"},
	"波兰": {"波兰", "PL", "Poland", "🇵🇱"},
	"瑞士": {"瑞士", "CH", "Switzerland", "🇨🇭"},
	"荷兰": {"荷兰", "NL", "Netherlands", "🇳🇱"},
	"乌克兰": {"乌克兰", "UA", "Ukraine", "🇺🇦"},
}

func DetectRegion(node *subscription.Node) string {
	name := node.Name
	for region, keywords := range regionKeywords {
		for _, kw := range keywords {
			if strings.Contains(name, kw) {
				return region
			}
		}
	}
	return "其他"
}

func GroupByRegion(nodes []*subscription.Node) map[string][]*subscription.Node {
	groups := make(map[string][]*subscription.Node)
	for _, node := range nodes {
		region := DetectRegion(node)
		groups[region] = append(groups[region], node)
	}
	return groups
}

func RegionOrder(groups map[string][]*subscription.Node) []string {
	order := []string{}
	seen := map[string]bool{}
	priority := []string{"香港", "台湾", "日本", "新加坡", "韩国", "美国", "英国", "德国", "法国", "澳大利亚", "加拿大", "俄罗斯", "印度", "巴西", "土耳其", "马来西亚", "泰国", "越南", "菲律宾", "印尼", "意大利", "智利", "尼日利亚", "澳门", "波兰", "瑞士", "荷兰", "乌克兰"}
	for _, r := range priority {
		if _, ok := groups[r]; ok {
			order = append(order, r)
			seen[r] = true
		}
	}
	for r := range groups {
		if !seen[r] {
			order = append(order, r)
		}
	}
	return order
}

func PromptRegion(groups map[string][]*subscription.Node) string {
	order := RegionOrder(groups)

	fmt.Println("\nAvailable regions:")
	for i, region := range order {
		nodes := groups[region]
		fmt.Printf("  %2d. %s (%d nodes)\n", i+1, region, len(nodes))
	}
	fmt.Printf("  %2d. All regions\n", len(order)+1)

	fmt.Print("\nSelect region number: ")
	var input string
	fmt.Scanln(&input)

	choice := -1
	for _, c := range strings.Fields(input) {
		fmt.Sscanf(c, "%d", &choice)
		break
	}

	if choice < 1 || choice > len(order)+1 {
		fmt.Fprintf(os.Stderr, "Invalid choice, using all regions\n")
		return ""
	}
	if choice == len(order)+1 {
		return ""
	}
	return order[choice-1]
}

func DisplayName(name string) string {
	cleaned := strings.Map(func(r rune) rune {
		if r == utf8.RuneError {
			return -1
		}
		return r
	}, name)
	return cleaned
}
