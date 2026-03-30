package ssh

import "sync"

const maxBufSize = 1024 * 1024

type RingBuffer struct {
	mu   sync.Mutex
	data []byte
}

func (r *RingBuffer) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data = append(r.data, p...)
	if len(r.data) > maxBufSize {
		r.data = r.data[len(r.data)-maxBufSize:]
	}
	return len(p), nil
}

func (r *RingBuffer) Snapshot() []byte {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]byte, len(r.data))
	copy(out, r.data)
	return out
}
