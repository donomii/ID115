package main


// To compile: export GODEBUG=cgocheck=0

import (
	"os"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"
	"github.com/willf/pad"

	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
)

var done = make(chan struct{})
var peripheralID string
var message string
var name string
var discovery bool
var spam bool

func onStateChanged(d gatt.Device, s gatt.State) {
	fmt.Println("State:", s)
	switch s {
	case gatt.StatePoweredOn:
		fmt.Println("Scanning...")
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	if (strings.ToUpper(p.ID()) == peripheralID) || (p.Name() == name) || (a.LocalName == name){
		// Stop scanning once we've got the peripheral we're looking for.
		p.Device().StopScanning()
		p.Device().Connect(p)
	}

	fmt.Printf("Peripheral ID:%s, NAME:(%s), ", p.ID(), p.Name())
	fmt.Println("Local Name =", a.LocalName)
	//fmt.Println("  TX Power Level    =", a.TxPowerLevel)
	//fmt.Println("  Manufacturer Data =", a.ManufacturerData)
	//fmt.Println("  Service Data      =", a.ServiceData)
	//fmt.Println("")

}

func onPeriphConnected(p gatt.Peripheral, err error) {
	fmt.Println("Connected")
	defer p.Device().CancelConnection(p)

	if err := p.SetMTU(500); err != nil {
		fmt.Printf("Failed to set MTU, err: %s\n", err)
	}

	// Discovery services
	ss, err := p.DiscoverServices(nil)
	if err != nil {
		fmt.Printf("Failed to discover services, err: %s\n", err)
		return
	}

	var targetChara *gatt.Characteristic

	for _, s := range ss {
		msg := "Service: " + s.UUID().String()
		if len(s.Name()) > 0 {
			msg += " (" + s.Name() + ")"
		}
		fmt.Println(msg)

		// Discovery characteristics
		cs, err := p.DiscoverCharacteristics(nil, s)
		if err != nil {
			fmt.Printf("Failed to discover characteristics, err: %s\n", err)
			continue
		}

		for _, c := range cs {
			msg := "  Characteristic  " + c.UUID().String()
			if len(c.Name()) > 0 {
				msg += " (" + c.Name() + ")"
			}
			msg += "\n    properties    " + c.Properties().String()
			fmt.Println(msg)
			if c.UUID().String() =="0af6" {
				fmt.Println("Found it!")
				targetChara = c
			}

			// Read the characteristic, if possible.
			if (c.Properties() & gatt.CharRead) != 0 {
				b, err := p.ReadCharacteristic(c)
				if err != nil {
					fmt.Printf("Failed to read characteristic, err: %s\n", err)
					continue
				}
				fmt.Printf("    value         %x | %q\n", b, b)
			}

			// Discovery descriptors
			ds, err := p.DiscoverDescriptors(nil, c)
			if err != nil {
				fmt.Printf("Failed to discover descriptors, err: %s\n", err)
				continue
			}

			for _, d := range ds {
				msg := "  Descriptor      " + d.UUID().String()
				if len(d.Name()) > 0 {
					msg += " (" + d.Name() + ")"
				}
				fmt.Println(msg)

				// Read descriptor (could fail, if it's not readable)
				b, err := p.ReadDescriptor(d)
				if err != nil {
					fmt.Printf("Failed to read descriptor, err: %s\n", err)
					continue
				}
				fmt.Printf("    value         %x | %q\n", b, b)
			}

			// Subscribe the characteristic, if possible.
			if (c.Properties() & (gatt.CharNotify | gatt.CharIndicate)) != 0 {
				f := func(c *gatt.Characteristic, b []byte, err error) {
					fmt.Printf("notified: % X | %q\n", b, b)
				}
				if err := p.SetNotifyValue(c, f); err != nil {
					fmt.Printf("Failed to subscribe characteristic, err: %s\n", err)
					continue
				}
			}

		}
		fmt.Println()
	}

	if targetChara != nil {
		myBytes := []byte(message)
		fmt.Println("Sending notification", myBytes,"!")
	
		p.WriteCharacteristic(targetChara, append([]byte{5, 3, 1, 1, 1, 4, 0, 8},  myBytes...), true)
		fmt.Println("Notification sent!")
	}

	//Note that we *must* wait for notifications, disconnecting immediately can cause the message we just wrote to be dropped
	fmt.Printf("Waiting for 5 seconds to get some notifications, if any.\n")
	time.Sleep(5*time.Second)
	if !spam {
		close(done)
	}
	fmt.Println("Peripheral probe complete")
}

func onPeriphDisconnected(p gatt.Peripheral, err error) {
	fmt.Println("Disconnected")
	if spam {
		p.Device().Scan(nil, true)
		p.Device().Scan([]gatt.UUID{}, false)
	} else {
		close(done)
	}
}

func main() {
	fmt.Println(`

********************************************************************

If you can't find your watch, check to see if it is connected
to another program, and disconnect it there.  MacOSX automatically
takes control of every Bluetooth LE device near it

********************************************************************

`)
	peripheralIDs := flag.String("id", "", "ID of device to notify (get from --discover)")
	names := flag.String("name", "", "Send to every device with this name")
	messages := flag.String("text", "", "Message to send")
	discovers := flag.Bool("discover", false, "Scan for devices")
	flag.BoolVar(&spam, "spam", false, "Notify every matching device")
	flag.Parse()
	fmt.Println("Spam state: ", spam)
	//if *peripheralIDs=="" {
		//log.Fatalf("Peripheral ID must be given")
	//}
	discovery = *discovers
	peripheralID= strings.ToUpper(*peripheralIDs)
	message = pad.Right(*messages, 12, " ")
	message = message[0:12]
	name = *names
	fmt.Println("Sending message: |", message, "|")

	if spam {
		go func () {
			time.Sleep(90*time.Second)
			fmt.Println("Killed by timer")
			os.Exit(0)
		}()
	}


	d, err := gatt.NewDevice(option.DefaultClientOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return
	}

	// Register handlers.
	d.Handle(
		gatt.PeripheralDiscovered(onPeriphDiscovered),
		gatt.PeripheralConnected(onPeriphConnected),
		gatt.PeripheralDisconnected(onPeriphDisconnected),
	)

	d.Init(onStateChanged)
	<-done
	fmt.Println("Done")
}
