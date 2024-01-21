// Package refcount provides a utility to help objects track the number of
// open instances, so they can release resources when not needed, and reopen
// on demand.
//
// TODO: Allow attaching an optional monitor to keep an eye on pathological
// patterns like thrashing.
//
// Antipated usage:
//
//	type Object struct {
//		rc *refcount.RefCount
//		path string
//		// ...
//	}
//
//	// New constructs a new Object, reading data from the given filesystem path.
//	func New(path string) (*Object, io.Closer, error) {
//		obj := &Object{
//			path: path,
//			// ...
//		}
//		obj.rc = refcount.New(obj.open, obj.close)
//		closer, err := obj.rc.Increment()
//		if err != nil {
//			return nil, nil, err
//		}
//		return obj, closer, nil
//	}
//
//	// Hold keeps the object open through multiple accesses, to avoid thrashing.
//	func (obj *Object) Hold() (io.Closer, error) {
//		return obj.rc.Increment()
//	}
//
//	// Get retrieves a value from this object. Call Close on the returned Closer
//	// when you are done with it.
//	func (obj *Object) Get(key keys.ResKey) (*Value, io.Closer, error) {
//		closer, err := obj.rc.Increment()
//		if err != nil {
//			return nil, nil, err
//		}
//		return obj.get(key), closer, nil
//	}
//
//	// close is called by the RefCount when the last open instance is closed.
//	func (obj *Object) close() error {
//		// Release any open resources.
//	}
//
//	// open is called by the RefCount when Increment is called on a closed
//	// object.
//	func (obj *Object) open() error {
//		// Open and load resources.
//	}
package refcount

import (
	"io"
	"runtime"
	"sync"
)

// Opener is called when Increment is called on a RefCount that had 0 open
// instances.
//
// Openers and Closers are protected against concurrent execution. Do not call
// Increment or Close from an Opener or a Closer.
type Opener func() error

// Closer is called when the last open instance's Close is called.
//
// Openers and Closers are protected against concurrent execution. Do not call
// Increment or Close from an Opener or a Closer.
type Closer func() error

// RefCount is the main type exported by this package.
//
// Hold one of these in your object, with the Opener set to acquire your
// object's heavy resources, and the Closer set to release those resources.
//
// Call Increment from a method that returns a close-able accessor from your
// object. It will open your resource if required.
//
// Call Decrement from the closer of that accessor object. It will close your
// resource if there are no more open instances.
type RefCount struct {
	// Needs to be synchronized so SetFinalizer will work.
	lock      sync.Mutex
	instances int

	opener Opener
	closer Closer
}

// New returns a new RefCount. This is meant to be an internal component of
// another object, not to be seen by users of your API.
func New(opener Opener, closer Closer) *RefCount {
	return &RefCount{
		opener: opener,
		closer: closer,
	}
}

// Instances returns the number of open instances. If this is greater than zero,
// this is the number of `Close` calls that would be needed (without any
// `Increment`s) to cause your object to actually be closed.
func (rc *RefCount) Instances() int {
	rc.lock.Lock()
	defer rc.lock.Unlock()
	return rc.instances
}

// Increment increments the number of open instances, calling the opener if
// necessary.
//
// When the instance is done being used, call `Close` on the returned
// `io.Closer`. Additional calls to `Close` beyond the first are successful
// no-ops.
func (rc *RefCount) Increment() (io.Closer, error) {
	rc.lock.Lock()
	defer rc.lock.Unlock()
	if rc.instances == 0 {
		err := rc.opener()
		if err != nil {
			return nil, err
		}
	}
	rc.instances++
	return rc.newCloser()
}

type decrementer struct {
	lock sync.Mutex
	rc   *RefCount
}

func (rc *RefCount) newCloser() (io.Closer, error) {
	// TODO: If there is a thrashing monitor, let it know if the finalizer
	// actually closes the object.
	dec := &decrementer{rc: rc}
	runtime.SetFinalizer(dec, func(dec *decrementer) error { return dec.Close() })
	return dec, nil
}

// Close decrements the number of open instances, calling the closer if
// necessary.
//
// Additional calls beyond the first are no-ops.
func (d *decrementer) Close() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.rc == nil {
		return nil // Already closed.
	}

	d.rc.lock.Lock()
	defer d.rc.lock.Unlock()
	if d.rc.instances == 1 {
		err := d.rc.closer()
		if err != nil {
			return err // Don't decrement; allow retry.
		}
	}

	runtime.SetFinalizer(d, nil)
	d.rc.instances--
	d.rc = nil // Prevent multiple decrement.
	return nil
}
