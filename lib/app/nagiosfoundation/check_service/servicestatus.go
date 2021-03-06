package nagiosfoundation

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

type serviceInfo struct {
	// The name of the service to process.
	desiredName string

	// The state of the service to match.
	desiredState string

	// The user of the service to match.
	desiredUser string

	actualName  string
	actualState string
	actualUser  string

	getServiceInfo func(string) (string, string, string, error)
}

// Returns the actual name of the service resulting from the service query.
func (i *serviceInfo) ActualName() string {
	return i.actualName
}

// Returns the actual state of the service resulting from the service query.
func (i *serviceInfo) ActualState() string {
	return i.actualState
}

// Returns the actual user of the service resulting from the service query.
func (i *serviceInfo) ActualUser() string {
	return i.actualUser
}

// Checks for a match against the actual name of the service. The comparison
// is case insensitive.
func (i *serviceInfo) IsName(name string) bool {
	return strings.EqualFold(i.ActualName(), name)
}

// Checks for a match against the actual state of the service. The comparison
// is case insensitive.
func (i *serviceInfo) IsState(state string) bool {
	return strings.EqualFold(i.ActualState(), state)
}

// Checks for a match against the actual user of the service. The comparison
// is case insensitive.
func (i *serviceInfo) IsUser(user string) bool {
	return strings.EqualFold(i.ActualUser(), user)
}

// Executes the OS constrained function to retrieve information about a service.
// This information is derived differently in Windows and Linux and must execute
// an OS constrained method named getInfoOsConstrained().
func (i *serviceInfo) GetInfo() error {
	var err error

	if i.getServiceInfo == nil {
		return errors.New("No get service info handler declared")
	}

	i.actualName, i.actualUser, i.actualState, err = i.getServiceInfo(i.desiredName)

	return err
}

// Process the desired service info against the actual service info and return
// check text and a return code.
func (i *serviceInfo) ProcessInfo() (string, int) {
	var checkInfo string
	var retcode int

	if !i.IsName(i.desiredName) {
		checkInfo = fmt.Sprintf("%s does not exist", i.desiredName)
		retcode = 2
	} else if i.desiredState != "" && i.desiredUser != "" {
		if i.IsState(i.desiredState) && i.IsUser(i.desiredUser) {
			checkInfo = fmt.Sprintf("%s in a %s state and started by user %s",
				i.ActualName(), i.ActualState(), i.ActualUser())
			retcode = 0
		} else {
			checkInfo = fmt.Sprintf("%s either not in a %s state or not started by user %s",
				i.ActualName(), i.desiredState, i.desiredUser)
			retcode = 2
		}
	} else if i.desiredState != "" {
		if i.IsState(i.desiredState) {
			checkInfo = fmt.Sprintf("%s in a %s state",
				i.ActualName(), i.ActualState())
			retcode = 0
		} else {
			checkInfo = fmt.Sprintf("%s not in a %s state",
				i.ActualName(), i.desiredState)
			retcode = 2
		}
	} else if i.desiredUser != "" {
		if i.IsUser(i.desiredUser) {
			checkInfo = fmt.Sprintf("%s started by user %s",
				i.ActualName(), i.ActualUser())
			retcode = 0
		} else {
			checkInfo = fmt.Sprintf("%s not started by user %s",
				i.ActualName(), i.desiredUser)
			retcode = 2
		}
	} else {
		checkInfo = fmt.Sprintf("%s in a %s state and started by user %s",
			i.ActualName(), i.ActualState(), i.ActualUser())
		retcode = 0
	}

	var responseStateText string
	var actualInfo string

	if retcode == 0 {
		responseStateText = "OK"
		actualInfo = ""
	} else {
		responseStateText = "CRITICAL"
		actualInfo = fmt.Sprintf(" (Name: %s, State: %s, User: %s)",
			i.ActualName(), i.ActualState(), i.ActualUser())
	}

	msg := fmt.Sprintf("CheckService %s - %s%s", responseStateText, checkInfo, actualInfo)
	return msg, retcode
}

// Show operating system independent help.
func showHelp() {
	fmt.Printf(
		`check_service -name <service name> [ other options ]
  Perform various checks for a service. These checks depend on the options given and
  the -name option is always required.

    -name <service name>: Required. The name of the service to check
`)
	showHelpOsConstrained()
}

// Check on a service. The executed functionality depends on the
// command line arguments given. The help text gives a good
// explanation of the arguments.
func CheckService() {
	if len(os.Args) < 2 {
		showHelp()
		os.Exit(0)
	}

	var name = flag.String("name", "", "the name of the service to check")
	var state = flag.String("state", "", "the desired state of the service")
	var user = flag.String("user", "", "the user the service should run as")
	var manager = flag.String("manager", "", "the name of local service manager. Allowed options are: systemd")
	flag.Parse()

	//fmt.Println("Checking Name:", *name, "State:", *state, "User:", *user, "Manager:", *manager)

	checkServiceOsConstrained(*name, *state, *user, *manager)
}
