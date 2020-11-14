package supervisor

import "fmt"

type supRefResponse interface {
	response()
}

type OK struct {}
func (*OK) response() {}

type ChildrenCount struct {
	// the total count of children, dead or alive
	Specs int
	// the count of all actively running child processes managed by this supervisor
	Active int
	// the count of all children marked as child_type = supervisor in the specification list,
	// regardless if the child process is still alive
	Supervisors int
	// the count of all children marked as child_type = worker in the specification list,
	// regardless if the child process is still alive
	Workers int
}
func (*ChildrenCount) response() {}
func (info *ChildrenCount) String() string {
	return fmt.Sprintf("----supervisor's children----\n" +
		"\t- all: 			%d\n" +
		"\t- active: 		%d\n" +
		"\t- supervisors: 	%d\n" +
		"\t- workers: 		%d\n",
		info.Specs, info.Active, info.Supervisors, info.Workers)
}
