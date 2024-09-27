package bs

// chatgpt wrote all of this, i'm using this as notes for what i want to build

import (
	"encoding/json"
	"errors"
	"fmt"
)

// bs is the struct that contains methods for OpenAIRequest and DeepQuery.
type bs struct {
	cache map[string]json.RawMessage // Cache to store retrieved context if needed
}

// OpenAIRequest is a placeholder for the method that interacts with the OpenAI API.
func (b *bs) OpenAIRequest(prompt string) (string, error) {
	// This function sends the prompt to OpenAI and returns the JSON response as a string.
	// Placeholder implementation. Replace with actual API call logic.
	return `{"possible_answer": "Example answer", "confidence": 0.85, "additional_information_needed": ["recent_interactions"], "suspected_humor": 0.1, "sentiment_analysis": 0.5, "facts_for_subsequent_queries": {}}`, nil
}

// DeepQuery simulates a map-reduce operation to fetch additional information based on the given input,
// performs state pruning, and returns the updated context as JSON.
func (b *bs) DeepQuery(infoList []string) (json.RawMessage, error) {
	// Simulate retrieving data from the storage layer (map phase)
	retrievedData := b.mapFunction(infoList)

	// Aggregate the retrieved data (reduce phase)
	finalContext := b.reduceFunction(retrievedData)

	// Prune outdated or irrelevant data from the state
	prunedContext := b.pruneState(finalContext)

	// Write back the pruned context to the storage layer (optional, based on requirements)
	err := b.writeBackToStorage(prunedContext)
	if err != nil {
		return nil, fmt.Errorf("error writing to storage: %w", err)
	}

	// Convert pruned context to JSON for returning to recurseQuery
	updatedState, err := json.Marshal(prunedContext)
	if err != nil {
		return nil, fmt.Errorf("error marshalling updated state: %w", err)
	}

	return json.RawMessage(updatedState), nil
}

// mapFunction simulates the retrieval of data based on the provided information list.
func (b *bs) mapFunction(infoList []string) []map[string]interface{} {
	// Placeholder logic: For each info in infoList, simulate retrieval of corresponding data.
	var retrievedData []map[string]interface{}
	for _, info := range infoList {
		// Simulate a generic data retrieval process (e.g., querying a database or object store)
		data := map[string]interface{}{
			info: fmt.Sprintf("Additional information related to %s", info),
		}
		retrievedData = append(retrievedData, data)
	}
	return retrievedData
}

// recurseQuery calls OpenAI's API recursively until the confidence level is >= 0.90 or the recursion depth limit is reached.
func (b *bs) recurseQuery(state json.RawMessage, depth int) (json.RawMessage, error) {
	// Set a reasonable recursion depth limit
	const maxDepth = 10
	if depth > maxDepth {
		return state, fmt.Errorf("maximum recursion depth reached")
	}

	// Convert the state JSON to a string prompt
	prompt := string(state)

	// Call the OpenAI API
	response, err := b.OpenAIRequest(prompt)
	if err != nil {
		return nil, fmt.Errorf("error calling OpenAI API: %w", err)
	}

	// Parse the response into a map to check the confidence level
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("error unmarshalling OpenAI response: %w", err)
	}

	// Check if the confidence field exists and is a float
	confidence, ok := result["confidence"].(float64)
	if !ok {
		return nil, errors.New("invalid response format: missing or incorrect 'confidence' field")
	}

	// Check if confidence level is sufficient
	if confidence >= 0.90 {
		// If confidence is high enough, return the result as JSON
		return json.RawMessage(response), nil
	}

	// Otherwise, attempt to get additional information using DeepQuery
	additionalInfo, _ := result["additional_information_needed"].([]interface{})
	var infoList []string
	for _, info := range additionalInfo {
		if infoStr, ok := info.(string); ok {
			infoList = append(infoList, infoStr)
		}
	}

	// If there are additional queries to be made, batch them in a single DeepQuery call
	if len(infoList) > 0 {
		additionalContext, err := b.DeepQuery(infoList)
		if err != nil {
			return nil, fmt.Errorf("error in DeepQuery: %w", err)
		}

		// Update the state JSON with the additional context
		var currentState map[string]interface{}
		if err := json.Unmarshal(state, &currentState); err != nil {
			return nil, fmt.Errorf("error unmarshalling state: %w", err)
		}

		// Assume additional context is a key-value pair and merge it into the current state
		var additionalData map[string]interface{}
		if err := json.Unmarshal(additionalContext, &additionalData); err != nil {
			return nil, fmt.Errorf("error unmarshalling additional context: %w", err)
		}

		for key, value := range additionalData {
			currentState[key] = value
		}

		// Convert updated state back to JSON
		updatedState, err := json.Marshal(currentState)
		if err != nil {
			return nil, fmt.Errorf("error marshalling updated state: %w", err)
		}

		// Update the state for the next recursion
		state = json.RawMessage(updatedState)
	}

	// Recursively call recurseQuery with the updated state
	return b.recurseQuery(state, depth+1)
}

// helpers

// reduceFunction aggregates the retrieved data into a single context.
func (b *bs) reduceFunction(retrievedData []map[string]interface{}) map[string]interface{} {
	finalContext := make(map[string]interface{})

	for _, data := range retrievedData {
		for key, value := range data {
			// Aggregate data into the final context; simple merging in this example
			finalContext[key] = value
		}
	}

	return finalContext
}

// pruneState simulates the pruning of outdated or irrelevant information from the context.
func (b *bs) pruneState(context map[string]interface{}) map[string]interface{} {
	// Placeholder pruning logic: This is where you can add rules for removing outdated information.
	// For this generic example, we assume all context is relevant and do nothing.
	return context
}

// writeBackToStorage simulates writing the pruned context back to the storage layer.
func (b *bs) writeBackToStorage(context map[string]interface{}) error {
	// Placeholder for actual storage logic (e.g., writing to an AWS S3 bucket).
	// This function writes the pruned state back to the storage layer.
	fmt.Printf("Writing updated context to storage: %+v\n", context)
	return nil
}
