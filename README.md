Currently just a bit of code to interact with iOS' Lockdownd and usbmuxd.

To run:

`go run lockdownclient.go usbmux.go`

You probably want to use libimobiledevice to setup the connection to the device: run `ssh -L 62078:/var/run/usbmuxd yourmachine` to forward the usbmuxd socket from the machine attached to the device to your dev machine.

The pairing record for your device should be placed in pairrecord.plist. You can find these in /var/lib/lockdown . In the future this'll grab these from usbmuxd automatically.

Licensed under the three-clause BSD license.
