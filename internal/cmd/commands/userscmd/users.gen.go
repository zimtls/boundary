// Code generated by "make cli"; DO NOT EDIT.
package userscmd

import (
	"errors"
	"fmt"
	"sync"

	"github.com/hashicorp/boundary/api"
	"github.com/hashicorp/boundary/api/users"
	"github.com/hashicorp/boundary/internal/cmd/base"
	"github.com/hashicorp/boundary/internal/cmd/common"
	"github.com/hashicorp/boundary/sdk/strutil"
	"github.com/mitchellh/cli"
	"github.com/posener/complete"
)

func initFlags() {
	flagsOnce.Do(func() {
		extraFlags := extraActionsFlagsMapFunc()
		for k, v := range extraFlags {
			flagsMap[k] = append(flagsMap[k], v...)
		}
	})
}

var (
	_ cli.Command             = (*Command)(nil)
	_ cli.CommandAutocomplete = (*Command)(nil)
)

type Command struct {
	*base.Command

	Func string

	plural string

	extraCmdVars
}

func (c *Command) AutocompleteArgs() complete.Predictor {
	initFlags()
	return complete.PredictAnything
}

func (c *Command) AutocompleteFlags() complete.Flags {
	initFlags()
	return c.Flags().Completions()
}

func (c *Command) Synopsis() string {
	if extra := extraSynopsisFunc(c); extra != "" {
		return extra
	}

	synopsisStr := "user"

	return common.SynopsisFunc(c.Func, synopsisStr)
}

func (c *Command) Help() string {
	initFlags()

	var helpStr string
	helpMap := common.HelpMap("user")

	switch c.Func {

	case "create":
		helpStr = helpMap[c.Func]() + c.Flags().Help()

	case "read":
		helpStr = helpMap[c.Func]() + c.Flags().Help()

	case "update":
		helpStr = helpMap[c.Func]() + c.Flags().Help()

	case "delete":
		helpStr = helpMap[c.Func]() + c.Flags().Help()

	case "list":
		helpStr = helpMap[c.Func]() + c.Flags().Help()

	default:

		helpStr = c.extraHelpFunc(helpMap)

	}

	// Keep linter from complaining if we don't actually generate code using it
	_ = helpMap
	return helpStr
}

var flagsMap = map[string][]string{

	"create": {"scope-id", "name", "description"},

	"read": {"id"},

	"update": {"id", "name", "description", "version"},

	"delete": {"id"},

	"list": {"scope-id", "filter", "recursive"},
}

func (c *Command) Flags() *base.FlagSets {
	if len(flagsMap[c.Func]) == 0 {
		return c.FlagSet(base.FlagSetNone)
	}

	set := c.FlagSet(base.FlagSetHTTP | base.FlagSetClient | base.FlagSetOutputFormat)
	f := set.NewFlagSet("Command Options")
	common.PopulateCommonFlags(c.Command, f, "user", flagsMap, c.Func)

	extraFlagsFunc(c, set, f)

	return set
}

func (c *Command) Run(args []string) int {
	initFlags()

	switch c.Func {
	case "":
		return cli.RunResultHelp
	}

	c.plural = "user"
	switch c.Func {
	case "list":
		c.plural = "users"
	}

	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.PrintCliError(err)
		return base.CommandUserError
	}

	if strutil.StrListContains(flagsMap[c.Func], "id") && c.FlagId == "" {
		c.PrintCliError(errors.New("ID is required but not passed in via -id"))
		return base.CommandUserError
	}

	var opts []users.Option

	if strutil.StrListContains(flagsMap[c.Func], "scope-id") {
		switch c.Func {

		case "create":
			if c.FlagScopeId == "" {
				c.PrintCliError(errors.New("Scope ID must be passed in via -scope-id or BOUNDARY_SCOPE_ID"))
				return base.CommandUserError
			}

		case "list":
			if c.FlagScopeId == "" {
				c.PrintCliError(errors.New("Scope ID must be passed in via -scope-id or BOUNDARY_SCOPE_ID"))
				return base.CommandUserError
			}

		}
	}

	client, err := c.Client()
	if err != nil {
		c.PrintCliError(fmt.Errorf("Error creating API client: %s", err.Error()))
		return base.CommandCliError
	}
	usersClient := users.NewClient(client)

	switch c.FlagName {
	case "":
	case "null":
		opts = append(opts, users.DefaultName())
	default:
		opts = append(opts, users.WithName(c.FlagName))
	}

	switch c.FlagDescription {
	case "":
	case "null":
		opts = append(opts, users.DefaultDescription())
	default:
		opts = append(opts, users.WithDescription(c.FlagDescription))
	}

	switch c.FlagRecursive {
	case true:
		opts = append(opts, users.WithRecursive(true))
	}

	if c.FlagFilter != "" {
		opts = append(opts, users.WithFilter(c.FlagFilter))
	}

	var version uint32

	switch c.Func {

	case "update":
		switch c.FlagVersion {
		case 0:
			opts = append(opts, users.WithAutomaticVersioning(true))
		default:
			version = uint32(c.FlagVersion)
		}

	case "add-accounts":
		switch c.FlagVersion {
		case 0:
			opts = append(opts, users.WithAutomaticVersioning(true))
		default:
			version = uint32(c.FlagVersion)
		}

	case "remove-accounts":
		switch c.FlagVersion {
		case 0:
			opts = append(opts, users.WithAutomaticVersioning(true))
		default:
			version = uint32(c.FlagVersion)
		}

	case "set-accounts":
		switch c.FlagVersion {
		case 0:
			opts = append(opts, users.WithAutomaticVersioning(true))
		default:
			version = uint32(c.FlagVersion)
		}

	}

	if ok := extraFlagsHandlingFunc(c, f, &opts); !ok {
		return base.CommandUserError
	}

	var result api.GenericResult

	var listResult api.GenericListResult

	switch c.Func {

	case "create":
		result, err = usersClient.Create(c.Context, c.FlagScopeId, opts...)

	case "read":
		result, err = usersClient.Read(c.Context, c.FlagId, opts...)

	case "update":
		result, err = usersClient.Update(c.Context, c.FlagId, version, opts...)

	case "delete":
		result, err = usersClient.Delete(c.Context, c.FlagId, opts...)

	case "list":
		listResult, err = usersClient.List(c.Context, c.FlagScopeId, opts...)

	}

	result, err = executeExtraActions(c, result, err, usersClient, version, opts)

	if err != nil {
		if apiErr := api.AsServerError(err); apiErr != nil {
			var opts []base.Option

			c.PrintApiError(apiErr, fmt.Sprintf("Error from controller when performing %s on %s", c.Func, c.plural), opts...)
			return base.CommandApiError
		}
		c.PrintCliError(fmt.Errorf("Error trying to %s %s: %s", c.Func, c.plural, err.Error()))
		return base.CommandCliError
	}

	output, err := printCustomActionOutput(c)
	if err != nil {
		c.PrintCliError(err)
		return base.CommandUserError
	}
	if output {
		return base.CommandSuccess
	}

	switch c.Func {

	case "delete":
		switch base.Format(c.UI) {
		case "json":
			if ok := c.PrintJsonItem(result); !ok {
				return base.CommandCliError
			}

		case "table":
			c.UI.Output("The delete operation completed successfully.")
		}

		return base.CommandSuccess

	case "list":
		switch base.Format(c.UI) {
		case "json":
			if ok := c.PrintJsonItems(listResult); !ok {
				return base.CommandCliError
			}

		case "table":
			listedItems := listResult.GetItems().([]*users.User)
			c.UI.Output(c.printListTable(listedItems))
		}

		return base.CommandSuccess

	}

	switch base.Format(c.UI) {
	case "table":
		c.UI.Output(printItemTable(result))

	case "json":
		if ok := c.PrintJsonItem(result); !ok {
			return base.CommandCliError
		}
	}

	return base.CommandSuccess
}

var (
	flagsOnce = new(sync.Once)

	extraActionsFlagsMapFunc = func() map[string][]string { return nil }
	extraSynopsisFunc        = func(*Command) string { return "" }
	extraFlagsFunc           = func(*Command, *base.FlagSets, *base.FlagSet) {}
	extraFlagsHandlingFunc   = func(*Command, *base.FlagSets, *[]users.Option) bool { return true }
	executeExtraActions      = func(_ *Command, inResult api.GenericResult, inErr error, _ *users.Client, _ uint32, _ []users.Option) (api.GenericResult, error) {
		return inResult, inErr
	}
	printCustomActionOutput = func(*Command) (bool, error) { return false, nil }
)
