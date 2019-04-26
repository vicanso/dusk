// Copyright 2019 tree xie
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dusk

type (
	// Instance dusk instance
	Instance struct {
		requestEvents  []*RequestEvent
		responseEvent  []*ResponseEvent
		errorListeners []ErrorListener
		doneListeners  []DoneListener
	}
)

// NewInstance new instance
func NewInstance() *Instance {
	return &Instance{}
}

// AddRequestListener add request listener
func (ins *Instance) AddRequestListener(ln RequestListener, eventType int) *Instance {
	if ins.requestEvents == nil {
		ins.requestEvents = make([]*RequestEvent, 0)
	}
	ins.requestEvents = append(ins.requestEvents, &RequestEvent{
		ln: ln,
		t:  eventType,
	})
	return ins
}

// AddResponseListener add response listener
func (ins *Instance) AddResponseListener(ln ResponseListener, eventType int) *Instance {
	if ins.responseEvent == nil {
		ins.responseEvent = make([]*ResponseEvent, 0)
	}
	ins.responseEvent = append(ins.responseEvent, &ResponseEvent{
		ln: ln,
		t:  eventType,
	})
	return ins
}

// AddErrorListener add error listener
func (ins *Instance) AddErrorListener(ln ErrorListener) *Instance {
	if ins.errorListeners == nil {
		ins.errorListeners = make([]ErrorListener, 0)
	}
	ins.errorListeners = append(ins.errorListeners, ln)
	return ins
}

// AddDoneListener add done listener
func (ins *Instance) AddDoneListener(ln DoneListener) *Instance {
	if ins.doneListeners == nil {
		ins.doneListeners = make([]DoneListener, 0)
	}
	ins.doneListeners = append(ins.doneListeners, ln)
	return ins
}

func (ins *Instance) attatchEvents(d *Dusk) {
	if ins.requestEvents != nil {
		d.addRequestEvent(ins.requestEvents...)
	}
	if ins.responseEvent != nil {
		d.addResponseEvent(ins.responseEvent...)
	}
	if ins.errorListeners != nil {
		d.AddErrorListener(ins.errorListeners...)
	}
	if ins.doneListeners != nil {
		d.AddDoneListener(ins.doneListeners...)
	}
}

// Get http get request
func (ins *Instance) Get(url string) *Dusk {
	d := Get(url)
	ins.attatchEvents(d)
	return d
}

// Head http head request
func (ins *Instance) Head(url string) *Dusk {
	d := Head(url)
	ins.attatchEvents(d)
	return d
}

// Post http post request
func (ins *Instance) Post(url string) *Dusk {
	d := Post(url)
	ins.attatchEvents(d)
	return d
}

// Put http put request
func (ins *Instance) Put(url string) *Dusk {
	d := Put(url)
	ins.attatchEvents(d)
	return d
}

// Patch http patch request
func (ins *Instance) Patch(url string) *Dusk {
	d := Patch(url)
	ins.attatchEvents(d)
	return d
}

// Delete http delete request
func (ins *Instance) Delete(url string) *Dusk {
	d := Delete(url)
	ins.attatchEvents(d)
	return d
}
