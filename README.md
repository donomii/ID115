# WARNING

## This program could destroy your watch.  Do not use it!

## But if you really want to anyway...

    ID115 --text="Hello world!" --id 382b0c42127c4532b259874bdbf41c4c


Where do you get the id (382b0c42127c4532b259874bdbf41c4c) from?  By running discover.

    ID115 --discover

but it is easier just to use the watch name

    ID115 --text="Hello world!" --name ID115

### Suggested uses

	make && ID115 --text "Job Done!" --name ID115

## Help, I can't find my watch!

Only one program per computer can connect to a watch at the same time (wtf).  Check that your default bluetooth manager hasn't connected to it by mistake.

## TO DO

* Add option to spam every watch within range
* Implement longer messages (requires fiddling with the multi-message vendor format)
* Probe devices in parallel?
* Notify all devices that match some criteria
