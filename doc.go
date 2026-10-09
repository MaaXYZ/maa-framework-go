// Package maa provides Go bindings for the MaaFramework.
//
// # Pipeline model
//
// Typed pipeline builders use pipeline v2 JSON with nested action and
// recognition objects. Decoding also accepts the legacy flat pipeline format
// and normalizes it into the v2 model; encoding always emits v2.
// For pipeline protocol details, see:
// https://github.com/MaaXYZ/MaaFramework/blob/main/docs/en_us/3.1-PipelineProtocol.md
//
// # Object lifetime
//
// Constructors such as NewTasker, NewResource, and the controller constructors
// return objects that own their native handles. Getters and callbacks —
// Tasker.GetResource, Tasker.GetController, Context.GetTasker, and the views
// passed to event callbacks — return borrowed views that do not own the
// native object; destroying one returns ErrBorrowed.
//
// Repeated successful Destroy calls on an owner are safe. Destroy returns
// ErrInUse while a call or asynchronous job is active, even if the returned
// Job was discarded; retry after it finishes. Call Wait on a job before
// destroying its owner if you need the job's outcome. A successful Destroy
// returns after native cleanup and prevents subsequent user callbacks;
// destroying an owner from one of its own callbacks returns ErrInCallback.
// After closing, methods that return an error report ErrClosed, and jobs
// expose it through Job.Error.
//
// Keep every resource and controller bound to a tasker alive until the tasker
// is destroyed, including bindings replaced by a later rebind; closing one
// before then returns ErrBound. Rebinding a running tasker returns
// ErrTaskerRunning. An AgentClient likewise retains its bound resource and
// registered event sources until the client is destroyed. A Context received
// in a callback, including a Clone, expires when the callback returns.
//
// # Concurrency
//
// Job and TaskJob support concurrent Wait, status queries and predicates, and
// Error. Multiple waiters share the completed result, and status queries
// remain available while a wait is in progress. Do not copy these objects.
//
// Most other operations require caller coordination: serialize option
// changes, resource and pipeline changes, calls through a single Context, and
// AgentClient and AgentServer lifecycle operations. Mutable configuration
// values such as Pipeline and Node also require caller synchronization.
// Handle lifetime checks do not make all native operations thread-safe.
//
// # Callbacks
//
// Callbacks execute synchronously on native calling threads and may overlap.
// Synchronize shared state in handlers, and never wait for work that needs
// the current callback to return. Do not call AgentServer lifecycle methods
// from its callbacks.
//
// # Registrations
//
// Change event sinks and custom recognition and action registrations only
// while the instance and all associated taskers are idle: stop new
// submissions and let their work and callbacks finish before adding,
// replacing, removing, or clearing registrations, and never change
// registrations from a callback. Configuration transactions on one instance
// are serialized with each other, including through borrowed views, but they
// must not overlap native execution.
package maa
