package bs

// chatgpt wrote all of this, i'm using this as notes for what i want to build

import (
	"encoding/json"
	"errors"
	"fmt"
	driver "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"github.com/google/uuid"
	"regexp"
	"time"
)

// bs is the struct that contains methods for OpenAIRequest and DeepQuery.
type bs struct {
	cache map[string]json.RawMessage // Cache to store retrieved context if needed
}

// initializeSession attempts to extract user ID and session ID from the prompt,
// retrieves the context if they exist, or creates a new session if not.
func (b *bs) initializeSession(userPrompt string) (string, string, *EnhancedContext, error) {
	var userID, sessionID string
	var currentContext *EnhancedContext

	// Try to extract user name and session ID using a regular expression
	re := regexp.MustCompile(`my name is (\w+) and the last sessionid i had was ([0-9a-fA-F-]+)`)
	matches := re.FindStringSubmatch(userPrompt)

	if len(matches) == 3 {
		// Extract user ID and session ID from the prompt
		userID = matches[1]
		sessionID = matches[2]

		// Attempt to retrieve the context from Neptune using DeepQuery (simulated here)
		retrievedRawData, err := b.DeepQuery([]string{sessionID})
		if err != nil {
			return "", "", nil, fmt.Errorf("error retrieving context from backend: %w", err)
		}

		// Parse the JSON raw message into a slice of maps
		var retrievedData []map[string]interface{}
		if err := json.Unmarshal(retrievedRawData, &retrievedData); err != nil {
			return "", "", nil, fmt.Errorf("error unmarshalling retrieved data: %w", err)
		}

		if len(retrievedData) > 0 {
			// Reduce the retrieved data to construct the enhanced context
			currentContext, err = b.reduceFunction(retrievedData)
			if err != nil {
				return "", "", nil, fmt.Errorf("error processing retrieved data: %w", err)
			}
		} else {
			// If no context is found, treat it as a new session
			sessionID = uuid.NewString()
			currentContext = NewEnhancedContext(userID, sessionID)
		}
	} else {
		// No user/session info found in prompt; create a new session
		userID = "newUser" // Default user ID; you might want to improve this logic
		sessionID = uuid.NewString()
		currentContext = NewEnhancedContext(userID, sessionID)
	}

	// Update the context with the initial prompt
	currentContext.History = append(currentContext.History, map[string]interface{}{
		"prompt":    userPrompt,
		"timestamp": time.Now().Format(time.RFC3339),
	})
	currentContext.Metadata["changeHistory"] = append(
		currentContext.Metadata["changeHistory"].([]string),
		fmt.Sprintf("Session initialized: %s", time.Now().Format(time.RFC3339)),
	)
	currentContext.LastUpdated = time.Now().Format(time.RFC3339)

	return userID, sessionID, currentContext, nil
}

// promptHandler processes a user's prompt, initializes or updates the context, and optionally queries the backend.
func (b *bs) promptHandler(userID, sessionID, userPrompt string, currentContext *EnhancedContext) (*EnhancedContext, error) {
	// Check if there's an existing context for this user session
	if currentContext == nil {
		currentContext = NewEnhancedContext(userID, sessionID)
		currentContext.Metadata["changeHistory"] = append(
			currentContext.Metadata["changeHistory"].([]string),
			fmt.Sprintf("New context created: %s", time.Now().Format(time.RFC3339)),
		)
	}

	// Add the new prompt to the context's history
	currentContext.History = append(currentContext.History, map[string]interface{}{
		"prompt":    userPrompt,
		"timestamp": time.Now().Format(time.RFC3339),
	})

	// Update the metadata to reflect the latest interaction
	currentContext.LastUpdated = time.Now().Format(time.RFC3339)
	currentContext.Metadata["changeHistory"] = append(
		currentContext.Metadata["changeHistory"].([]string),
		fmt.Sprintf("Prompt added to history: %s", time.Now().Format(time.RFC3339)),
	)

	// Determine if a backend query is needed based on the user's prompt or current state
	infoList := []string{userPrompt} // You might extract specific keywords or entities to query
	additionalContext, err := b.DeepQuery(infoList)
	if err != nil {
		return nil, fmt.Errorf("error querying backend: %w", err)
	}

	// Parse the additionalContext into a map if needed
	var additionalData map[string]interface{}
	if err := json.Unmarshal(additionalContext, &additionalData); err != nil {
		return nil, fmt.Errorf("error unmarshalling additional context: %w", err)
	}

	// Integrate the results from additionalContext into the current context's Metadata
	for key, value := range additionalData {
		switch v := value.(type) {
		case string:
			currentContext.Metadata[key] = v
		case int, float64, bool:
			currentContext.Metadata[key] = fmt.Sprintf("%v", v)
		case map[string]interface{}, []interface{}:
			jsonValue, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("error marshalling value for key %s: %w", key, err)
			}
			currentContext.Metadata[key] = string(jsonValue)
		default:
			// Skip unsupported types
			continue
		}
	}

	// Additional processing can be added here, such as passing the updated context to OpenAI

	return currentContext, nil
}

// OpenAIRequest is a placeholder for the method that interacts with the OpenAI API.
func (b *bs) OpenAIRequest(prompt string) (string, error) {
	// This function sends the prompt to OpenAI and returns the JSON response as a string.
	// Placeholder implementation. Replace with actual API call logic.
	return `{"possible_answer": "Example answer", "confidence": 0.85, "additional_information_needed": ["recent_interactions"], "suspected_humor": 0.1, "sentiment_analysis": 0.5, "facts_for_subsequent_queries": {}}`, nil
}

func (b *bs) DeepQuery(infoList []string) (json.RawMessage, error) {
	// 1. Retrieve data from the graph using the infoList (map phase)
	retrievedData, err := b.mapFunction(infoList)
	if err != nil {
		return nil, fmt.Errorf("error retrieving data: %w", err)
	}

	// 2. Aggregate the retrieved data (reduce phase)
	finalContext, err := b.reduceFunction(retrievedData)
	if err != nil {
		return nil, fmt.Errorf("error reducing data: %w", err)
	}

	// 3. Prune outdated or irrelevant data from the state
	prunedContext, err := b.pruneState(finalContext)
	if err != nil {
		return nil, fmt.Errorf("error pruning state: %w", err)
	}

	// 4. Write back the pruned context to the graph (optional)
	err = b.writeBackToStorage(prunedContext)
	if err != nil {
		return nil, fmt.Errorf("error writing to storage: %w", err)
	}

	// 5. Convert pruned context to JSON for returning
	updatedState, err := json.Marshal(prunedContext)
	if err != nil {
		return nil, fmt.Errorf("error marshalling updated state: %w", err)
	}

	return json.RawMessage(updatedState), nil
}

// Example implementation of mapFunction for Neptune using Gremlin
func (b *bs) mapFunction(infoList []string) ([]map[string]interface{}, error) {
	// Create a new Gremlin client
	client, err := driver.NewClient("ws://your-neptune-endpoint:8182/gremlin")
	if err != nil {
		return nil, fmt.Errorf("error creating Gremlin client: %w", err)
	}
	defer client.Close() // Ensure the client is closed when done

	var retrievedData []map[string]interface{}

	for _, info := range infoList {
		// Construct the Gremlin query
		query := fmt.Sprintf("g.V().has('property', '%s').valueMap()", info)

		// Execute the query
		resultSet, err := client.Submit(query)
		if err != nil {
			return nil, fmt.Errorf("error executing Gremlin query: %w", err)
		}

		// Use All() to gather all results into a slice
		results, err := resultSet.All()
		if err != nil {
			return nil, fmt.Errorf("error retrieving results: %w", err)
		}

		// Iterate over the results
		for _, result := range results {
			// Extract the value (which is a map) from the result
			data, ok := result.GetInterface().(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("unexpected result format: %v", result.GetInterface())
			}

			retrievedData = append(retrievedData, data)
		}
	}

	return retrievedData, nil
}

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
		additionalContext, err := b.DeepQuery(infoList) // This will interact with Neptune
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

// reduceFunction aggregates the retrieved data into a single context.
func (b *bs) reduceFunction(retrievedData []map[string]interface{}) (*EnhancedContext, error) {
	finalContext := &EnhancedContext{
		History:         []map[string]interface{}{},
		Entities:        []string{},
		Concepts:        []string{},
		OpenAIParams:    map[string]interface{}{},
		UserPreferences: map[string]interface{}{},
		Metadata: map[string]interface{}{
			"changeHistory": []string{},
		},
	}

	for _, data := range retrievedData {
		for key, value := range data {
			switch key {
			case "history":
				// Append to the history list
				if history, ok := value.([]map[string]interface{}); ok {
					finalContext.History = append(finalContext.History, history...)
				} else {
					return nil, fmt.Errorf("unexpected type for history: %T", value)
				}

			case "entities":
				// Merge entities (remove duplicates)
				if entities, ok := value.([]string); ok {
					finalContext.Entities = mergeStringSlices(finalContext.Entities, entities)
				} else {
					return nil, fmt.Errorf("unexpected type for entities: %T", value)
				}

			case "concepts":
				// Merge concepts (remove duplicates)
				if concepts, ok := value.([]string); ok {
					finalContext.Concepts = mergeStringSlices(finalContext.Concepts, concepts)
				} else {
					return nil, fmt.Errorf("unexpected type for concepts: %T", value)
				}

			case "openAIParams":
				// Merge OpenAI parameters
				if params, ok := value.(map[string]interface{}); ok {
					for pKey, pValue := range params {
						finalContext.OpenAIParams[pKey] = pValue
					}
				} else {
					return nil, fmt.Errorf("unexpected type for openAIParams: %T", value)
				}

			case "userPreferences":
				// Merge user preferences
				if preferences, ok := value.(map[string]interface{}); ok {
					for prefKey, prefValue := range preferences {
						finalContext.UserPreferences[prefKey] = prefValue
					}
				} else {
					return nil, fmt.Errorf("unexpected type for userPreferences: %T", value)
				}

			case "metadata":
				// Merge metadata
				if metadata, ok := value.(map[string]interface{}); ok {
					mergedMetadata, err := mergeValues(finalContext.Metadata, metadata)
					if err != nil {
						return nil, fmt.Errorf("error merging metadata: %w", err)
					}
					finalContext.Metadata = mergedMetadata.(map[string]interface{})
				} else {
					return nil, fmt.Errorf("unexpected type for metadata: %T", value)
				}

			default:
				// For other keys, add them to metadata
				finalContext.Metadata[key] = value
			}
		}
	}

	return finalContext, nil
}

// Helper function to merge two slices of strings, removing duplicates.
func mergeStringSlices(slice1, slice2 []string) []string {
	uniqueMap := make(map[string]bool)
	for _, item := range slice1 {
		uniqueMap[item] = true
	}
	for _, item := range slice2 {
		uniqueMap[item] = true
	}

	mergedSlice := []string{}
	for key := range uniqueMap {
		mergedSlice = append(mergedSlice, key)
	}

	return mergedSlice
}

// Helper function to handle merging of conflicting values
func mergeValues(existingValue, newValue interface{}) (interface{}, error) {
	// Example: If both values are slices, combine them
	switch existing := existingValue.(type) {
	case []interface{}:
		if newSlice, ok := newValue.([]interface{}); ok {
			return append(existing, newSlice...), nil
		}
		return nil, fmt.Errorf("type mismatch: expected []interface{}, got %T", newValue)
	// Handle other types as needed (e.g., maps, strings, etc.)
	default:
		// Simple case: Overwrite with the new value
		return newValue, nil
	}
}

// pruneState simulates the pruning of outdated or irrelevant information from the context.
func (b *bs) pruneState(context *EnhancedContext) (*EnhancedContext, error) {
	// Convert the EnhancedContext to JSON for OpenAI processing
	contextJSON, err := json.Marshal(context)
	if err != nil {
		return nil, fmt.Errorf("error marshalling context: %w", err)
	}

	// Construct the prompt for OpenAI
	prompt := fmt.Sprintf(`Here is the current context along with metadata:
%s

Identify elements that are outdated or irrelevant based on their last updated timestamp, change history, or any other indicators in the context. Provide the updated, pruned context as JSON:`, string(contextJSON))

	// Call OpenAI to get the pruned context
	response, err := b.OpenAIRequest(prompt)
	if err != nil {
		return nil, fmt.Errorf("error calling OpenAI API: %w", err)
	}

	// Parse the response from OpenAI to get the pruned context
	var prunedContext EnhancedContext
	if err := json.Unmarshal([]byte(response), &prunedContext); err != nil {
		return nil, fmt.Errorf("error unmarshalling OpenAI response: %w", err)
	}

	// Update metadata to reflect the pruning action
	prunedContext.Metadata["changeHistory"] = append(
		prunedContext.Metadata["changeHistory"].([]string),
		fmt.Sprintf("Context pruned: %s", time.Now().Format(time.RFC3339)),
	)
	prunedContext.LastUpdated = time.Now().Format(time.RFC3339)

	return &prunedContext, nil
}

// writeBackToStorage simulates writing the pruned context back to the storage layer.
func (b *bs) writeBackToStorage(context *EnhancedContext) error {
	// Create a new Gremlin client
	client, err := driver.NewClient("ws://your-neptune-endpoint:8182/gremlin")
	if err != nil {
		return fmt.Errorf("error creating Gremlin client: %w", err)
	}
	defer client.Close()

	// Update or create the vertex for the session using its session ID
	query := fmt.Sprintf("g.V().has('sessionID', '%s').fold().coalesce(unfold(), addV('session').property('sessionID', '%s'))", context.SessionID, context.SessionID)
	_, err = client.Submit(query)
	if err != nil {
		return fmt.Errorf("error executing Gremlin query for session: %w", err)
	}

	// Write back key fields in the context
	err = b.updateContextProperties(client, context)
	if err != nil {
		return fmt.Errorf("error updating context properties: %w", err)
	}

	return nil
}

// updateContextProperties updates properties in the storage based on the EnhancedContext fields.
func (b *bs) updateContextProperties(client *driver.Client, context *EnhancedContext) error {
	// Example: Write back the metadata
	metadataJSON, err := json.Marshal(context.Metadata)
	if err != nil {
		return fmt.Errorf("error marshalling metadata: %w", err)
	}
	query := fmt.Sprintf("g.V().has('sessionID', '%s').property('metadata', '%s')", context.SessionID, string(metadataJSON))
	_, err = client.Submit(query)
	if err != nil {
		return fmt.Errorf("error updating metadata: %w", err)
	}

	// Example: Write back the history (store as JSON string)
	historyJSON, err := json.Marshal(context.History)
	if err != nil {
		return fmt.Errorf("error marshalling history: %w", err)
	}
	query = fmt.Sprintf("g.V().has('sessionID', '%s').property('history', '%s')", context.SessionID, string(historyJSON))
	_, err = client.Submit(query)
	if err != nil {
		return fmt.Errorf("error updating history: %w", err)
	}

	// Write other fields as needed (e.g., entities, concepts)
	entitiesJSON, err := json.Marshal(context.Entities)
	if err != nil {
		return fmt.Errorf("error marshalling entities: %w", err)
	}
	query = fmt.Sprintf("g.V().has('sessionID', '%s').property('entities', '%s')", context.SessionID, string(entitiesJSON))
	_, err = client.Submit(query)
	if err != nil {
		return fmt.Errorf("error updating entities: %w", err)
	}

	conceptsJSON, err := json.Marshal(context.Concepts)
	if err != nil {
		return fmt.Errorf("error marshalling concepts: %w", err)
	}
	query = fmt.Sprintf("g.V().has('sessionID', '%s').property('concepts', '%s')", context.SessionID, string(conceptsJSON))
	_, err = client.Submit(query)
	if err != nil {
		return fmt.Errorf("error updating concepts: %w", err)
	}

	return nil
}
