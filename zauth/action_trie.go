package zauth

import (
	"context"
	"github.com/bytedance/sonic"
	"github.com/zohu/zgin/zch"
	"sort"
	"strings"
)

func LoadPermission(ctx context.Context, uid string) *PermTrie {
	key := zch.PrefixAuthAction.Key(uid)
	per := zch.R().Get(ctx, key).Val()
	if per != "" {
		zch.R().Expire(ctx, key, options.Age)
	}
	return BuildPermissionTrieFromString(per)
}
func SavePermission(ctx context.Context, uid string, patterns []string) {
	key := zch.PrefixAuthAction.Key(uid)
	per := BuildPermissionTrie(patterns)
	zch.R().Set(ctx, key, per.String(), options.Age)
}

type PermTrie struct {
	Root  *TrieNode `json:"root"`
	Allow bool      `json:"allow"` // *:*:*
}
type TrieNode struct {
	Children map[string]*TrieNode `json:"children"`
	IsLeaf   bool                 `json:"is_leaf"`
	Level    int                  `json:"level"`
}

// BuildPermissionTrie
// @Description: 构建权限树
// @param patterns
// @return *PermTrie
func BuildPermissionTrie(patterns []string) *PermTrie {
	patterns = mergePatterns(patterns)
	trie := &PermTrie{Root: newTrieNode(0)}
	for _, pattern := range patterns {
		// 检查全局通配符
		if pattern == "*:*:*" {
			trie.Allow = true
		}
		parts := strings.Split(pattern, ":")
		if len(parts) != 3 {
			continue
		}
		currentNode := trie.Root
		for i, part := range parts {
			// 创建或获取子节点
			if _, exists := currentNode.Children[part]; !exists {
				currentNode.Children[part] = newTrieNode(i)
			}
			currentNode = currentNode.Children[part]
			// 标记最后一段为叶子节点
			if i == 2 {
				currentNode.IsLeaf = true
			}
		}
	}
	return trie
}
func BuildPermissionTrieFromString(permission string) *PermTrie {
	trie := &PermTrie{Root: newTrieNode(0)}
	if permission == "" {
		return trie
	}
	_ = sonic.UnmarshalString(permission, trie)
	return trie
}

func newTrieNode(level int) *TrieNode {
	return &TrieNode{
		Children: make(map[string]*TrieNode),
		Level:    level,
	}
}
func (t *PermTrie) Match(action string) bool {
	// 先检查全局通配符
	if t.Allow {
		return true
	}
	parts := strings.Split(action, ":")
	if len(parts) != 3 {
		return false
	}
	return t.matchRecursive(t.Root, parts, 0)
}
func (t *PermTrie) String() string {
	str, _ := sonic.MarshalString(t)
	return str
}
func (t *PermTrie) matchRecursive(node *TrieNode, parts []string, depth int) bool {
	// 到达最后一段，检查是否叶子节点
	if depth == 3 {
		return node.IsLeaf
	}
	currentPart := parts[depth]
	// 尝试精确匹配
	if child, ok := node.Children[currentPart]; ok {
		if t.matchRecursive(child, parts, depth+1) {
			return true
		}
	}
	// 尝试通配符匹配
	if child, ok := node.Children["*"]; ok {
		if t.matchRecursive(child, parts, depth+1) {
			return true
		}
	}
	return false
}

func mergePatterns(patterns []string) []string {
	if len(patterns) == 0 {
		return patterns
	}
	unique := make(map[string]struct{})
	var cleaned []string
	for _, p := range patterns {
		if !isValidPattern(p) {
			continue
		}
		if _, exists := unique[p]; !exists {
			unique[p] = struct{}{}
			cleaned = append(cleaned, p)
		}
	}
	sort.Slice(cleaned, func(i, j int) bool {
		wi := wildcardCount(cleaned[i])
		wj := wildcardCount(cleaned[j])
		if wi != wj {
			return wi > wj
		}
		return cleaned[i] < cleaned[j]
	})
	merged := make(map[string]struct{})
	for i := 0; i < len(cleaned); i++ {
		pattern := cleaned[i]
		shouldMerge := false

		// 检查是否可以被已处理的模式覆盖
		for mp := range merged {
			if pattern == mp {
				shouldMerge = true
				break
			}
			if coversPattern(mp, pattern) {
				shouldMerge = true
				break
			}
		}

		// 未被任何已有模式覆盖则添加
		if !shouldMerge {
			merged[pattern] = struct{}{}
		}
	}
	result := make([]string, 0, len(merged))
	for p := range merged {
		result = append(result, p)
	}
	return result
}
func isValidPattern(pattern string) bool {
	parts := strings.Split(pattern, ":")
	if len(parts) != 3 {
		return false
	}
	for _, p := range parts {
		if p == "" || strings.ContainsAny(p, " \t\n\r") {
			return false
		}
	}
	return true
}
func wildcardCount(pattern string) int {
	parts := strings.Split(pattern, ":")
	count := 0
	for _, p := range parts {
		if p == "*" {
			count++
		}
	}
	return count
}
func coversPattern(a, b string) bool {
	partsA := strings.Split(a, ":")
	partsB := strings.Split(b, ":")

	for i := 0; i < 3; i++ {
		// A为通配符时匹配任何值
		if partsA[i] == "*" {
			continue
		}
		// B为通配符时要求A必须是通配符
		if partsB[i] == "*" {
			return false
		}
		// 两者都是具体值但不同
		if partsA[i] != partsB[i] {
			return false
		}
	}
	return true
}
