DISKIE(1)

# NAME

Diskie - CLI for udisks

# SYNOPSIS

*diskie* *print*  [OPTION...]++
*diskie* *select* [OPTION...] [--] MENU_CMD [MENU_ARG...]

*diskie* *mount*  [OPTION...] [--] DEVICE [ASKPASS_CMD [MENU_ARGS...]]++
*diskie* *attach* [OPTION...] [--] DEVICE [ASKPASS_CMD [MENU_ARGS...]]++
*diskie* *open*   [OPTION...] [--] DEVICE [ASKPASS_CMD [MENU_ARGS...]]

*diskie* *unmount* [OPTION...] [--] DEVICE++
*diskie* *detach*  [OPTION...] [--] DEVICE

# DESCRIPTION

*diskie* is a high-level frontend for *udisks*(8),
aiming to offer helpful features that *udisksctl*(1) does not.

See the EXAMPLES section below to get started quickly.

# OPTIONS

*-h*, *--help*
	Print a help message and exit.

*-v*, *--version*
	Print the version number and exit.

# COMMANDS

*print* [OPTION...]

	Print the available block devices.

	Options:

	*-f*, *--format*=FORMAT

		Output format.

		Possible values are:

		- json-array (Default)
		- json-map
		- tabular
		- basic
		- rofi-markup
		- template:FILE_PATH (e.g., template:~/template.txt)

		see the FORMATS section below for more info.

	*-l*, *--limit*=NUMBER

		NUMBER is an integer between 0 and 3,
		where higher values progressively filter out
		less significant devices.

		See the LIMIT section for more info.

		Defaults to 0.

*select* [OPTION...] [--] MENU_CMD [MENU_ARGS...]

	Select a device using a dmenu-compatible program.

	Options:

	*-f*, *--format*=FORMAT

		Output format.

		Possible values are:

		- basic (Default)
		- tabular
		- rofi-markup
		- template:FILE_PATH (e.g., template:~/template.txt)

		see the FORMATS section below for more info.

	*-l*, *--limit*=NUMBER

		NUMBER is an integer between 0 and 3,
		where higher values progressively filter out
		less significant devices.

		See the LIMIT section for more info.

		Defaults to 2.

	*-L*, *--menu-max-lines*=NUMBER

		Limit the maximum value of the %l sequence to NUMBER.

		See the MENU COMMAND section below for more info.

		Defaults to 0.

*mount*   [OPTION...] [--] DEVICE [MENU_CMD [MENU_ARGS...]]++
*attach*  [OPTION...] [--] DEVICE [MENU_CMD [MENU_ARGS...]]++
*open*    [OPTION...] [--] DEVICE [MENU_CMD [MENU_ARGS...]]++
*unmount* [OPTION...] [--] DEVICE++
*detach*  [OPTION...] [--] DEVICE

	Perform ACTION on DEVICE.

	Diskie requires a password to unlock an encrypted device.

	If the *--password-file* option is specified,
	the password is read from the file.

	Otherwise, if MENU_CMD is specified,
	it's executed and upon successful exit,
	the command's standard output is regarded as the password.

	If neither *--password-file* nor MENU_CMD are specified,
	the password will be prompted from the controlling terminal.
	
	For security reasons,
	there is no option to provide the password directly on the command line.

	Possible values for ACTION are:

	*mount*
		Mount the selected device if it's a filesystem.
		Print the mountpoint to standard output after a successful mount.

	*unmount*
		Unmount the selected device if it's already mounted.

	*attach*
		Mount the selected device if it's a filesystem.
		Attempt to unlock it first if it's an encrypted device.
		Print the mountpoint to standard output after a successful mount.

	*detach*
		Unmount the selected device if it's already mounted.
		Attempt to lock it after unmounting if it's an encrypted device.

	*open*
		Similar to *attach*,
		but open the mountpoint after mounting the device
		(using freedesktop's file-manager-interface)

	Options:

	*-p*, *--password-file*=FILE_PATH

		Read the password from the given file.

# FORMATS

*json-array*
	JSON array of block device objects.

*json-map*
	JSON object where the keys are udisks object paths,
	and the values are block device objects.

*basic*
	Basic non-tabularized newline-separated format.
	Suitable for use with dmenu, rofi, etc.

*tabular*
	Tabularized newline-separated format.
	Suitable for use with dmenu, rofi, etc.

*rofi-markup*
	Similar to basic,
	but with some aesthetic improvements
	which require rofi's *-markup-rows* option.

*template*:FILE_PATH
	Path to a file containing a Golang template.
	An initial tilde is expanded to the user's home directory.
	See the TEMPLATE section below for more info.

# LIMIT

Increasing the limit level
progressively hides additional devices
from the displayed list of devices,
in addition to those previously filtered:

*0*
	No devices are hidden from the list.

*1*
	Hide devices that neither contain a filesystem nor are encrypted
	(unmountable and unlockable devices)

*2*
	Hide devices that aren't user-facing
	(e.g., Windows "System Reserved" partitions)

*3*
	Hide non-removable devices,
	and devices whose media is unavailable
	(e.g., an empty DVD-RW drive).

# MENU COMMAND

MENU_CMD must be a dmenu-compatible (e.g., dmenu, rofi, fzf) command.
Choices are passed to the command's standard input.
If the command exits successfully,
it's standard output is regarded as the user's
selection.

If the menu does not exit successfully,
Diskie will exit with the same exit code as the menu.
This behavior is useful to detect rofi's kb-custom-N keys
which manipulate rofi's exit code based on the keybinding that is pressed.

The sequence *%l* in the command
is replaced by an integer representing the number of available choices.
This number is limited by the value of the *--menu-max-lines* option,
unless it is set to 0, in which case there is no limit.

# TEMPLATE

The template file should contain a Golang template
which defines the format of each output line.

See https://pkg.go.dev/text/template
for details on Golang's template specification.

All newlines are removed from the results of template execution.

Template functions from the library *Masterminds/sprig*[1]
are available for use in templates,
as well as template functions from the file *template-funcs.go*:

https://github.com/Masterminds/sprig/blob/@SPRIG_VERSION@/docs/index.md
https://github.com/koonix/diskie/blob/@LATEST_TAG@/cmd/diskie/template-funcs.go

templates of the default available formats (tabular, basic, ...)
are defined in the file *formats.go*
and can be utilized as examples:

https://github.com/koonix/diskie/blob/@LATEST_TAG@/cmd/diskie/formats.go

# EXAMPLES

====================

Mount /dev/sdc2 and open it's mountpoint on success:

```
diskie open /dev/sdc2
```

====================

Mount /dev/sda1
and print the mountpoint to stdout,
asking for a password using rofi if the device is encrypted:

```
diskie attach /dev/sda1 --
	rofi -dmenu -password -no-fixed-num-lines
```

====================

Print a list of removable devices:

```
diskie print -l3
```

====================

Ask for a device from a list of removable devices using fzf
and print it's info to stdout:

```
diskie select \\
	--limit 3 \\
	--format template:~/tmpl.txt \\
	fzf --prompt='Diskie > '
```

====================

Script that asks for a device to mount using dmenu.

```
#!/bin/bash

device=$(
	diskie select --limit=2 --format=tabular \\
		--menu-max-lines=20 -- dmenu -l %l -p Diskie
)

action=$(
	printf '%s\n' Open Detach | dmenu -p Diskie
)

case $action in
	Open) diskie open "$device" -- dmenu -p 'Diskie Password' ;;
	Detach) diskie detach "$device" ;;
esac
```

====================

Script that asks for a device to mount using rofi.
Pressing Ctrl+r in rofi reloads the list of devices.

```
#!/bin/bash

while :; do
	device=$(
		diskie select \\
			--limit=2 \\
			--format=rofi-markup \\
			--menu-max-lines=20 \\
			rofi -dmenu -markup-rows -l %l -p Diskie \\
			-kb-custom-1 control+r
	)
	case $? in
		0) break ;; # success
		10) ;; # control+r
		*) exit $? ;;
	esac
done

action=x
        wat
        piss
      fuck
      shit
  printf
  piss
x

case $action in
	Open)
		diskie open "$device" \\
			rofi -dmenu -password -no-fixed-num-lines -p 'Diskie Password'
	;;
	Detach)
		diskie detach "$device"
	;;
esac
```

# SEE ALSO

*udisks*(8), *udisksctl*(1)
