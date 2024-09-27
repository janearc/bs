**Prompt:**

This question is part of an application which uses an LLM to answer questions from humans in a conversational format. The function posing these questions may be called recursively if the confidence is not high enough, attempting to add context through questions to the end user until sufficiently high confidence is reached. For each question provided, respond in JSON. Your JSON response should include the following elements:

1. `"possible_answer"`: A possible answer to the question that you believe addresses the user's concern.
2. `"confidence"`: A number between 0 and 1 representing your confidence in the sufficiency of the answer you provided. 
   - **0** means no confidence in the response, 
   - **1** means very high confidence that the response is accurate and sufficient.
3. `"additional_information_needed"`: An array of guesses about what unknown things that, if known, might result in higher confidence in the answer.
4. `"suspected_humor"`: A number between 0 and 1 representing likelihood the original question is humorous. 
   - **0** means it is not humorous at all, 
   - **1** means it is highly humorous.
5. `"sentiment_analysis"`: A number between 0 and 1 representing the sentiment of the original question. 
   - **0** indicates very negative sentiment, 
   - **1** indicates very positive sentiment.
6. `"facts_for_subsequent_queries"`: A dictionary containing any facts or contextual elements that should be included in subsequent queries to ensure context is preserved as more facts are uncovered.

Here is an example question and the desired JSON format:

**Example Question:** "I have been having trouble scheduling my friends Danielle, Emily, and Katherine. I'm not sure how to encourage them to be more flexible or creative about possible times we could all meet without seeming like I am being overbearing or telling them what to do. How can we solve this scheduling problem?"

**Desired JSON Response Format:**

```json
{
  "possible_answer": "One approach could be to suggest a shared online calendar or scheduling tool where everyone can indicate their availability. This way, it feels more like a collaborative effort rather than you setting the schedule. You could also propose a few options and ask them to choose what works best, emphasizing that you're open to adjusting your own schedule to accommodate theirs. Encouraging a brainstorming session on times or formats for meetups might help spark some creative solutions without coming across as overbearing.",
  "confidence": 0.7,
  "additional_information_needed": [
    "Have you used any scheduling tools (e.g., Doodle, Google Calendar) with them before?",
    "What are their typical availability patterns (work hours, other commitments)?",
    "How have they responded to past scheduling attempts?",
    "Are there specific constraints (e.g., weekends only, evenings preferred) that need to be considered?",
    "Is there a preferred way to communicate with them about scheduling (text, email, group chat)?"
  ],
  "suspected_humor": 0.05,
  "sentiment_analysis": 0.3,
  "facts_for_subsequent_queries": {
    "people_involved": ["Danielle", "Emily", "Katherine"],
    "situation_description": "trouble scheduling a meetup",
    "desired_outcome": "find flexible or creative times to meet",
    "potential_constraints": "avoiding appearing overbearing"
  }
}
```

Answer must include these specific elements in JSON format, following the above guidelines.
