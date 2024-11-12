package main

// Request holds the input data
type Request struct {
	Name    string   `json:"name,require"`
	Include []string `json:"include"`
	Exclude []string `json:"exclude"`
	Inherits string `json:"inherits,omitempty"`
}

// Distributor holds the permissions for each distributor
type Distributor struct {
	Name    string
	Include []string
	Exclude []string
	Inherits string
}

var distributorDB = make(map[string]*Distributor)

// createDistributor creates a new distributor or clones from a parent distributor if inheritance is defined.
func (d *Request) createDistributor() *Distributor {
	if d.Inherits != "" {
		// If Inherits is set, we clone from the parent distributor
		parent := distributorDB[d.Inherits]
		return parent.clone(d)
	}
	// If no inheritance, create a new distributor
	return &Distributor{
		Name:    d.Name,
		Include: d.Include,
		Exclude: d.Exclude,
		Inherits: d.Inherits, // Inheritance remains unchanged
	}
}

// clone creates a new distributor based on the parent's permissions and merges the include/exclude lists
func (d *Distributor) clone(inData *Request) *Distributor {
	var includeList []string
	for _, code := range inData.Include {
		res,_ := checkDistributor(d.Name, code)
		if res == MsgAccessGranted {
			includeList = append(includeList, code)
		}
	}
	return &Distributor{
		Name:    inData.Name,
		Include: includeList,
		Exclude: append(d.Exclude, inData.Exclude...),
		Inherits: inData.Inherits, // Inheritance remains unchanged
	}
}
