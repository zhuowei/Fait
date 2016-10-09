Currently just a bit of code to interact with iOS' Lockdownd.

You probably want to use libimobiledevice to setup the connection to Lockdownd for now: run
`iproxy 62078 62078`

The pairing record for your device should be placed in pairrecord.plist. You can find these in /var/lib/lockdown . In the future this'll grab these from lockdownd automatically.

Licensed under the three-clause MIT license.
