# WARNING

## This program could destroy your watch.  Do not use it!

## But if you really want to anyway...

    ID115 --text="Hello world!" --id 382b0c42127c4532b259874bdbf41c4c --type ID115


Where do you get the id (382b0c42127c4532b259874bdbf41c4c) from?  By running discover.

    ID115 --discover

but it is easier just to use the watch name

    ID115 --text="Hello world!" --name ID115 --type ID115

### Options

	--text		The message to display
	--type		ID115 or HBand
	--id		The peripheral id
	--name		The "local name" to send to
	--verbose	Print LOTS of extra debugging

### Supported watches

ID115 and HBand

More to come!


## Suggested uses

	make && ID115 --text "Job Done!" --name ID115 --type ID115

## Help, I can't find my watch!

Only one program per computer can connect to a watch at the same time (wtf).  Check that your default bluetooth manager hasn't connected to it by mistake.

## Idiosyncracies

Lots.  Possibly the worst from a user point of view is that whenever another device unpairs the watch, the watch gets itself a new peripheral ID.  This also appears to confuse the android app.

This explains why so many customers are complaining that they are having troubles connecting to the watch, and why it appears to stop working for them.  It also appears that you can screw over a random stranger by pairing with their watch, then unpairing.

To make this worse, there are no further identifying features for the watch.  There are no details in the Manufacturers data section, or the Service data section.  Which means:

* There is no way to identify an individual watch
* There is no way to tell which version of hardware or software your watch has
* The only way to tell if it is an ID115 watch is to check the bluetooth name for "ID115"

This goes a long way towards explaining why there are 3 different versions of the android software for this line of watches, and why each version only works with some watches and not others.


## TO DO

* Add option to spam every watch within range
* Implement longer messages (requires fiddling with the multi-message vendor format) (done!)
* Probe devices in parallel?
* Notify all devices that match some criteria
* Add more functions (camera, alarm, set time, etc)

## Credits

This program is a heavy modification of one of the example programs in "github.com/paypal/gatt"
