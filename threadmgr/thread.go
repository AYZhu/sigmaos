package threadmgr

import (
	"sync"

	//	"github.com/sasha-s/go-deadlock"

	np "ulambda/ninep"
)

/*
 * The Thread struct ensures that only one operation on the thread is being run
 * at any one time. It assumes that any condition variables which are passed
 * into it are only being used/slept on by one goroutine. It relies on the
 * locking behavior of sessconds in the layer above it to ensure that there
 * aren't any sleep/wake races. Sessconds' locking behavior should be changed
 * with care.
 */

type Thread struct {
	sync.Locker
	newOpCond *sync.Cond // Signalled when there are new ops available.
	cond      *sync.Cond // Signalled when the currently running op is about to sleep or terminate. This indicates that the thread can continue running.
	done      bool
	ops       []*Op        // List of (new) ops to process.
	wakeups   []*sync.Cond // List of conds (goroutines/ops) that need to wake up.
	pfn       ProcessFn
}

func makeThread(pfn ProcessFn) *Thread {
	t := &Thread{}
	t.Locker = &sync.Mutex{}
	t.newOpCond = sync.NewCond(t.Locker)
	t.cond = sync.NewCond(t.Locker)
	t.ops = []*Op{}
	t.wakeups = []*sync.Cond{}
	t.pfn = pfn
	return t
}

// Enqueue a new op to be processed.
func (t *Thread) Process(fc *np.Fcall, replies chan *np.Fcall) {
	t.Lock()
	defer t.Unlock()

	t.ops = append(t.ops, makeOp(fc, replies))

	// Signal that there are new ops to be processed.
	t.newOpCond.Signal()
}

// Notify the Thread that a goroutine/operation would like to go to sleep.
func (t *Thread) Sleep(c *sync.Cond) {

	// Acquire the *t.Mutex to ensure that the Thread.run thread won't miss the
	// sleep signal.
	t.Lock()

	// Notify the thread that this goroutine is going to sleep.
	t.cond.Signal()

	// Allow the thread to make progress, as this goroutine is about to sleep.
	// Sleep/wake races here are avoided by SessCond locking above us.
	t.Unlock()

	// Sleep on this condition variable.
	c.Wait()
}

// Called when an operation is going to be woken up. The caller may be a
// goroutine/operation belonging to this thread/session or the
// goroutine/operation may belong to another thread/session.
func (t *Thread) Wake(c *sync.Cond) {
	t.Lock()
	defer t.Unlock()

	// Append to the list of goroutines to be woken up.
	t.wakeups = append(t.wakeups, c)

	// Notify Thread that there are more wakeups to be processed. This signal
	// will be missed if the thread is not currently waiting for new ops (in the
	// tight loop in Thread.run). However, this is ok, as it checks for any new
	// wakeups that need to be processed before calling newOpCond.Wait() again.
	t.newOpCond.Signal()
}

// Processes operations on a single channel.
func (t *Thread) run() {
	t.Lock()
	for {
		// Check if we have pending wakeups or new ops.
		for len(t.ops) == 0 && len(t.wakeups) == 0 {
			if t.done {
				return
			}

			// Wait until we have new wakeups or new ops.
			t.newOpCond.Wait()
		}

		// If there is a new op to process, process it.
		if len(t.ops) > 0 {
			// Get the next op.
			op := t.ops[0]
			t.ops = t.ops[1:]

			// Process the op. Run the next op in a new goroutine (which may sleep or
			// block).
			go func() {
				// Execute the op.
				t.pfn(op.fc, op.replies)
				// Lock to make sure the completion signal isn't missed.
				t.Lock()
				// Notify the Thread that the op has completed.
				t.cond.Signal()
				// Unlock to allow the Thread to make progress.
				t.Unlock()
			}()
			// Wait for the op to sleep or complete.
			t.cond.Wait()
		}
		// Process any pending wakeups. These may have been generated by the new op,
		// or by an op/goroutine belonging to another Thread in the system.
		t.processWakeups()
	}
}

// Processes wakeups.
func (t *Thread) processWakeups() {
	// Create a copy of the list of pending wakeups.
	tmp := []*sync.Cond{}
	for _, c := range t.wakeups {
		tmp = append(tmp, c)
	}
	// Empty the wakeups list.
	t.wakeups = t.wakeups[:0]
	// wake up each goroutine/op, then wait for it to complete or sleep again.
	for _, c := range tmp {
		// Wake up the sleeping goroutine.
		c.Signal()
		// Wait for the operation to sleep or terminate.
		t.cond.Wait()
	}
}

func (t *Thread) start() {
	go t.run()
}

func (t *Thread) stop() {
	t.Lock()
	defer t.Unlock()
	t.done = true
	t.cond.Signal()
}
