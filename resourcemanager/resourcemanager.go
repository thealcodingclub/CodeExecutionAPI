package resourcemanager

import "sync"

var (
	availableMemory int = 6000000 // 6GB in KB
	memoryMutex     sync.Mutex
)

func ReserveMemory(memory int) bool {
	memoryMutex.Lock()
	defer memoryMutex.Unlock()

	if availableMemory >= memory {
		availableMemory -= memory
		return true
	}
	return false
}

func ReleaseMemory(memory int) {
	memoryMutex.Lock()
	defer memoryMutex.Unlock()

	availableMemory += memory
}
