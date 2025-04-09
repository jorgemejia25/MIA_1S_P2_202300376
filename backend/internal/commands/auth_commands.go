package commands

import (
	"fmt"
	"strings"

	"disk.simulator.com/m/v2/internal/args"
	"disk.simulator.com/m/v2/internal/disk/operations/auth"
	"github.com/spf13/cobra"
)

var authRootCmd = &cobra.Command{Use: "auth"}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to the system",
	RunE: func(cmd *cobra.Command, args []string) error {
		user, _ := cmd.Flags().GetString("user")
		password, _ := cmd.Flags().GetString("pass")
		id, _ := cmd.Flags().GetString("id")

		if user == "" || password == "" || id == "" {
			return fmt.Errorf("user, password and id are required")
		}

		output := fmt.Sprintf("Logging in with user %s and id %s", user, id)

		fmt.Fprintln(cmd.OutOrStdout(), output)

		// Aquí iría la lógica para autenticar al usuario
		err := auth.Login(user, password, id)

		if err != nil {
			return err
		}

		return nil
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from the system",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Aquí iría la lógica para cerrar la sesión

		err := auth.Logout()

		if err != nil {
			return err
		}

		output := "Logged out"

		fmt.Fprintln(cmd.OutOrStdout(), output)

		return nil
	},
}

var mkgrpCmd = &cobra.Command{
	Use:   "mkgrp",
	Short: "Create a new group",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")

		err := auth.CreateGroup(name)

		if err != nil {
			return err
		}

		output := fmt.Sprintf("Group %s created", name)

		fmt.Fprintln(cmd.OutOrStdout(), output)

		return nil
	},
}

var mkusrCmd = &cobra.Command{
	Use:   "mkusr",
	Short: "Create a new user",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("user")
		password, _ := cmd.Flags().GetString("pass")
		group, _ := cmd.Flags().GetString("grp")

		err := auth.CreateUser(name, password, group)

		if err != nil {
			return err
		}

		output := fmt.Sprintf("User %s created", name)

		fmt.Fprintln(cmd.OutOrStdout(), output)

		return nil
	},
}

var rmgrpCmd = &cobra.Command{
	Use:   "rmgrp",
	Short: "Remove a group",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")

		err := auth.RemoveGroup(name)

		if err != nil {
			return err
		}

		output := fmt.Sprintf("Group %s removed", name)

		fmt.Fprintln(cmd.OutOrStdout(), output)

		return nil
	},
}

var rmusrCmd = &cobra.Command{
	Use:   "rmusr",
	Short: "Remove a user",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("user")

		err := auth.RemoveUser(name)

		if err != nil {
			return err
		}

		output := fmt.Sprintf("User %s removed", name)

		fmt.Fprintln(cmd.OutOrStdout(), output)

		return nil
	},
}

var chgrp = &cobra.Command{
	Use:   "chgrp",
	Short: "Change the group of a user",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("user")
		group, _ := cmd.Flags().GetString("grp")

		err := auth.ChangeGroup(name, group)

		if err != nil {
			return err
		}

		output := fmt.Sprintf("User %s changed to group %s", name, group)

		fmt.Fprintln(cmd.OutOrStdout(), output)

		return nil
	},
}

func init() {
	authRootCmd.AddCommand(loginCmd)
	// Login
	loginCmd.PersistentFlags().StringP("user", "u", "", "Username")
	loginCmd.MarkPersistentFlagRequired("user")
	loginCmd.PersistentFlags().StringP("pass", "p", "", "Password")
	loginCmd.MarkPersistentFlagRequired("pass")
	loginCmd.PersistentFlags().StringP("id", "i", "", "Partition ID")
	loginCmd.MarkPersistentFlagRequired("id")

	// Logout
	authRootCmd.AddCommand(logoutCmd)

	// Mkgrp
	authRootCmd.AddCommand(mkgrpCmd)
	mkgrpCmd.PersistentFlags().StringP("name", "n", "", "Group name")
	mkgrpCmd.MarkPersistentFlagRequired("name")

	// Mkusr
	authRootCmd.AddCommand(mkusrCmd)
	mkusrCmd.PersistentFlags().StringP("user", "n", "", "Username")
	mkusrCmd.MarkPersistentFlagRequired("user")
	mkusrCmd.PersistentFlags().StringP("pass", "p", "", "Password")
	mkusrCmd.MarkPersistentFlagRequired("pass")
	mkusrCmd.PersistentFlags().StringP("grp", "g", "", "Group")
	mkusrCmd.MarkPersistentFlagRequired("grp")

	// Rmgrp
	authRootCmd.AddCommand(rmgrpCmd)
	rmgrpCmd.PersistentFlags().StringP("name", "n", "", "Group name")
	rmgrpCmd.MarkPersistentFlagRequired("name")

	// Rmusr
	authRootCmd.AddCommand(rmusrCmd)
	rmusrCmd.PersistentFlags().StringP("user", "n", "", "Username")
	rmusrCmd.MarkPersistentFlagRequired("user")

	// Chgrp
	authRootCmd.AddCommand(chgrp)
	chgrp.PersistentFlags().StringP("user", "n", "", "Username")
	chgrp.MarkPersistentFlagRequired("user")
	chgrp.PersistentFlags().StringP("grp", "g", "", "Group")
	chgrp.MarkPersistentFlagRequired("grp")

}

func ParseAuthCommand(
	command string,
	data string,
) (
	string,
	error,
) {
	args := args.SplitArgs(data)

	// Parsear los argumentos
	authRootCmd.SetArgs(args)

	// Capturar la salida del comando
	output := &strings.Builder{}
	authRootCmd.SetOut(output)

	// Ejecutar el comando
	err := authRootCmd.Execute()
	if err != nil {
		return "", err
	}

	if len(authRootCmd.Flags().Args()) > 0 {
		return "", fmt.Errorf("unknown arguments: %v", authRootCmd.Flags().Args())
	}

	// Devolver la salida capturada
	return output.String(), nil
}
