package resourcemanager

var availableMemory int = 6000000 // 6GB in KB

func ReserveMemory(memory int) bool {
	if availableMemory >= memory {
		availableMemory -= memory
		return true
	}
	return false
}

func ReleaseMemory(memory int) {
	availableMemory += memory
}
