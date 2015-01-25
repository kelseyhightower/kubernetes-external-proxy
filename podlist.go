package main

type PodList struct {
	Items []Pod `json:"items"`
}

type Pod struct {
	ID           string       `json:"id"`
	CurrentState CurrentState `json:"currentState"`
}

type CurrentState struct {
	Status string `json:"status"`
	PodIP  string `json:"podIP"`
}
