package common

import (
	"fmt"
	"strings"
)

type ComplianceType int

const (
	ComplianceScan ComplianceType = iota
	ComplianceSuite
	ScanSettingBinding
	Profile
	Rule
	TypeUnkown
)

func ValidateObjectArgs(args []string) (objtype ComplianceType, objname string, err error) {

	if len(args) < 1 {
		err = fmt.Errorf("You need to specify at least one object")
		return
	}

	if len(args) > 2 {
		err = fmt.Errorf("unkown argument(s): %s", args[2:])
		return
	}

	var rawobjtype string
	if len(args) == 1 {
		objref := strings.Split(args[0], "/")
		if len(objref) == 1 {
			return TypeUnkown, "", fmt.Errorf("Missing object name")
		}

		if len(objref) > 2 {
			return TypeUnkown, "", fmt.Errorf("Malformed reference to object: %s", args[0])
		}

		rawobjtype = objref[0]
		objname = objref[1]
	} else {
		rawobjtype = args[0]
		objname = args[1]
	}

	objtype, err = GetValidObjType(rawobjtype)
	return
}

func GetValidObjType(rawtype string) (ComplianceType, error) {
	switch rawtype {
	case "ScanSettingBindings", "ScanSettingBinding", "scansettingbindings", "scansettingbinding":
		return ScanSettingBinding, nil
	case "ComplianceSuites", "ComplianceSuite", "compliancesuites", "compliancesuite":
		return ComplianceSuite, nil
	case "ComplianceScans", "ComplianceScan", "compliancescans", "compliancescan":
		return ComplianceScan, nil
	case "Profiles", "Profile", "profiles", "profile":
		return Profile, nil
	case "Rules", "Rule", "rules", "rule":
		return Rule, nil
	default:
		return TypeUnkown, fmt.Errorf("Unkown object type: %s", rawtype)
	}
}
