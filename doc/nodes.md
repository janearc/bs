# concept
this started as something I was working on inside of [sux](https://github.com/janearc/sux/issues/5) but the more i think about it the more it deserves to be fleshed out here, and this keeps sux smaller. sux should really be just a storage and retrieval layer.

# structures
```golang
type Interaction struct {
    Details        string `json:"details"`
    Date           string `json:"date"`
    SentimentScore float64 `json:"sentiment_score"`
}
```

```golang
type Actor struct {
    Name            string `json:"name"`
    BehaviorPatterns string `json:"behavior_patterns"`
    RelativeSentimentVector string `json:"sentiment_vector"`
}
```

```golang
type Query struct {
    Actors   []Actor `json:"actors"`
    SentimentScore float64 `json:"sentiment_score"`
    HumorScore float64  `json:"humor_score"`
    RawQuery string  `json:"raw_query"`
    AnswerConfidence float64  `json:"answer_confidence"`
    Gaps []GapQuery `json:"gaps"`
}
```
