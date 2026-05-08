package engine

import (
	"context"
	"strings"
)

type Rule struct {
	ID          string
	Platform    string
	Patterns    []string
	RootCause   string
	Suggestions []Suggestion
	Confidence  float64
}

type MatchRequest struct {
	Platform string
	ErrorMsg string
	TopK     int
}

type MatchResult struct {
	Matched    bool
	Rule       *Rule
	RootCause  string
	Suggestions []Suggestion
	Confidence float64
	Card       *KnowledgeCard
}

type KnowledgeBase interface {
	Retrieve(ctx context.Context, req *RetrieveRequest) ([]*KnowledgeCard, error)
}

type KnowledgeCard struct {
	ID             string
	Platform       string
	ErrorType      string
	ErrorPatterns  []string
	RootCause      string
	Suggestions    []Suggestion
	Confidence     float64
}

type RetrieveRequest struct {
	Platform string
	ErrorMsg string
	TopK     int
}

type RuleEngine struct {
	kb       KnowledgeBase
	rules    []Rule
	trie     *Trie
}

type Trie struct {
	root *TrieNode
}

type TrieNode struct {
	children map[rune]*TrieNode
	isEnd    bool
	rule     *Rule
}

func NewTrie() *Trie {
	return &Trie{root: &TrieNode{children: make(map[rune]*TrieNode)}}
}

func (t *Trie) Insert(pattern string, rule *Rule) {
	node := t.root
	pattern = strings.ToLower(pattern)
	for _, ch := range pattern {
		if _, ok := node.children[ch]; !ok {
			node.children[ch] = &TrieNode{children: make(map[rune]*TrieNode)}
		}
		node = node.children[ch]
	}
	node.isEnd = true
	node.rule = rule
}

func (t *Trie) Search(text string) *Rule {
	node := t.root
	text = strings.ToLower(text)
	for _, ch := range text {
		if _, ok := node.children[ch]; !ok {
			return nil
		}
		node = node.children[ch]
	}
	if node.isEnd {
		return node.rule
	}
	return nil
}

func NewRuleEngine(kb KnowledgeBase, rules []Rule) *RuleEngine {
	trie := NewTrie()
	for i := range rules {
		for _, pattern := range rules[i].Patterns {
			trie.Insert(pattern, &rules[i])
		}
	}
	return &RuleEngine{
		kb:   kb,
		rules: rules,
		trie:  trie,
	}
}

func (e *RuleEngine) Match(ctx context.Context, req *MatchRequest) *MatchResult {
	for _, rule := range e.rules {
		if rule.Platform != req.Platform {
			continue
		}
		for _, pattern := range rule.Patterns {
			if strings.Contains(strings.ToLower(req.ErrorMsg), strings.ToLower(pattern)) {
				return &MatchResult{
					Matched:     true,
					Rule:        &rule,
					RootCause:   rule.RootCause,
					Suggestions: rule.Suggestions,
					Confidence:  rule.Confidence,
				}
			}
		}
	}

	if e.kb != nil {
		cards, _ := e.kb.Retrieve(ctx, &RetrieveRequest{
			Platform: req.Platform,
			ErrorMsg: req.ErrorMsg,
			TopK:     1,
		})
		if len(cards) > 0 {
			return &MatchResult{
				Matched:    true,
				Card:      cards[0],
				RootCause:  cards[0].RootCause,
				Confidence: cards[0].Confidence,
			}
		}
	}

	return &MatchResult{Matched: false, Confidence: 0}
}

func (e *RuleEngine) AddRule(rule Rule) {
	e.rules = append(e.rules, rule)
	for _, pattern := range rule.Patterns {
		e.trie.Insert(pattern, &rule)
	}
}

func (e *RuleEngine) RuleCount() int {
	return len(e.rules)
}

func (e *RuleEngine) GetRulesByPlatform(platform string) []Rule {
	var result []Rule
	for _, rule := range e.rules {
		if rule.Platform == platform {
			result = append(result, rule)
		}
	}
	return result
}
