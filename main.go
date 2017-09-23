package main


// To compile: export GODEBUG=cgocheck=0

import (
	"syscall"
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
var rawMessage string
var name string
var discovery bool
var spam bool
var verbose bool
var bandType string

func onStateChanged(d gatt.Device, s gatt.State) {
	if verbose { fmt.Println("State:", s) }
	switch s {
	case gatt.StatePoweredOn:
		if verbose { fmt.Println("Scanning...") }
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	if (strings.ToUpper(p.ID()) == peripheralID) || ((name != "") && (p.Name() == name)) || ((name != "") && (a.LocalName == name)){
		// Stop scanning once we've got the peripheral we're looking for.
		if !discovery {
			p.Device().StopScanning()
			p.Device().Connect(p)
		}
		if verbose {
			fmt.Println("Found device, connecting...")
		}
	}

	fmt.Printf("Peripheral ID:%s, names:(%s, %s), \n", p.ID(), p.Name(), a.LocalName)
	if verbose {
		fmt.Println("Local Name =", a.LocalName)
		fmt.Println("  TX Power Level    =", a.TxPowerLevel)
		fmt.Println("  Manufacturer Data =", a.ManufacturerData)
		fmt.Println("  Service Data      =", a.ServiceData)
		fmt.Println("")
		fmt.Println("Found device, connecting...")
	}

}

func readCharacteristic(p gatt.Peripheral, c *gatt.Characteristic) {
	defer func() {
		if r := recover(); r != nil {
		    fmt.Println("Recovered in readChara", r)
		}
	}()

	b, err := p.ReadCharacteristic(c)
	if err != nil {
		fmt.Printf("Failed to read characteristic, err: %s\n", err)
		return
	}
	if verbose {
	fmt.Printf("    value         %x | %q\n", b, b)
	}
}

func readDescriptor(p gatt.Peripheral, d *gatt.Descriptor) {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("Recovered in readDes", r)
        }
    }()
// Read descriptor (could fail, if it's not readable)
	//fmt.Printf("%+v\n", d)
	b, err := p.ReadDescriptor(d)
	if err != nil {
		fmt.Printf("Failed to read descriptor, err: %s\n", err)
		return
	}
	if verbose {
		fmt.Printf("    value         %x | %q\n", b, b)
	}
}

func discoverDescriptorsWrapper(p gatt.Peripheral, c *gatt.Characteristic) (ds []*gatt.Descriptor, err error){
	defer func() {
		if r := recover(); r != nil {
		    fmt.Println("Recovered in DesWrap", r)
		}
	}()
	//err = errors.New("unknown error")
	ds, err = p.DiscoverDescriptors(nil, c)
	return
}


func discoverCharacteristicsWrapper(p gatt.Peripheral, s *gatt.Service) (cs []*gatt.Characteristic, err error){
	defer func() {
		if r := recover(); r != nil {
		    fmt.Println("Recovered in CharWrap", r)
		}
	}()
	//err = errors.New("unknown error")
	fmt.Printf("%+v\n", s)
	cs, err = p.DiscoverCharacteristics(nil, s)
	return
}



func onPeriphConnected(p gatt.Peripheral, err error) {
	fmt.Println("Connected")
	fmt.Printf("%+v\n", p)
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
	found := false

	for _, s := range ss {
		msg := "Service: " + s.UUID().String()
		if len(s.Name()) > 0 {
			msg += " (" + s.Name() + ")"
		}
		fmt.Println(msg)

		// Discovery characteristics
		cs, err := discoverCharacteristicsWrapper(p, s)
		if err != nil || cs == nil {
			fmt.Printf("Failed to discover characteristics, err: %s\n", err)
			continue
		}

		for _, c := range cs {
			msg := "  Characteristic  (" + fmt.Sprintf("%v",c.Handle()) + ")" + c.UUID().String()
			if len(c.Name()) > 0 {
				msg += " (" + c.Name() + ")"
			}
			msg += "\n    properties    " + c.Properties().String()
	if verbose {
			fmt.Println(msg)
	}
			if c.UUID().String() =="0af6" || c.UUID().String() == "f008000304514000b000000000000000" {
		if verbose {
				fmt.Println("Found it!", c.UUID().String())
		}
				targetChara = c
				found = true
			}

			// Read the characteristic, if possible.
			if (c.Properties() & gatt.CharRead) != 0 {
				readCharacteristic(p,c)
			}

			// Discovery descriptors
			ds, err := discoverDescriptorsWrapper(p, c)
			if err != nil {
				fmt.Printf("Failed to discover descriptors, err: %s\n", err)
				continue
			}

			for _, d := range ds {
				msg := "  Descriptor      ("+ fmt.Sprintf("%v",d.Handle()) +")" + d.UUID().String()
				if len(d.Name()) > 0 {
					msg += " (" + d.Name() + ")"
				}
		if verbose {
				fmt.Println(msg)
		}

				readDescriptor(p, d)
			}


		}
		fmt.Println()
	}

	if found {
		fmt.Println("Sending message to", targetChara.UUID().String()) 
		if bandType =="ID115" {
			message, message_length := padTo(rawMessage, 22)   //Two packets worth of data
			head := message[0:12]
			rest := message[13:22]
			head_length := 0
			rest_length := 0
			if message_length <13 {
				head_length = message_length
				rest_length = 0
			} else {
				head_length = 12
				rest_length = message_length-12
			}

			fmt.Println("Sending notification", head, rest,"!", head_length, rest_length)
		
			p.WriteCharacteristic(targetChara, append([]byte{5, 3, 2, 1, 1, byte(message_length), 0, 8},  head...), true)
			p.WriteCharacteristic(targetChara, append([]byte{5, 3, 2, 2 },  rest...), true)
			fmt.Println("Notification sent!")
		} else {
			display_mode := byte(6)
			message, message_length := padTo(rawMessage, 28)   //Two packets worth of data
			head := message[0:14]
			rest := message[15:28]
			head_length := 0
			rest_length := 0
			if message_length <15 {
				head_length = message_length
				rest_length = 0
			} else {
				head_length = 14
				rest_length = message_length-14
			}
			fmt.Println("Sending notification", head, ":", message,"!")
		
			packet := append([]byte{194, display_mode, byte(head_length), 2, 1, 2},  []byte(head)...)
			p.WriteCharacteristic(targetChara, packet, true)
			fmt.Println("Sent", packet," with payload of length ", head_length)

			packet = append([]byte{194, display_mode, byte(rest_length), 2, 2, 2},  []byte(rest)...)
			p.WriteCharacteristic(targetChara, packet, true)
		if verbose {
			fmt.Println("Sent", packet," with payload of length ", rest_length)
		}

			//Repeat, because that's what we have to do
			//packet = append([]byte{194, 17, byte(rest_length), 3, 3, 2},  []byte(rest)...)
			//p.WriteCharacteristic(targetChara, packet, true)
			//fmt.Println("Sent", packet," with payload of length ", rest_length)
			//time.Sleep(2*time.Second)

			//No idea what this does, but the official software does, so we do it too
			p.WriteCharacteristic(targetChara, []byte{193, 1, 1, 1}, true)
			fmt.Println("Sent", []byte{193, 1, 1, 1},"!")

			fmt.Println("Notification sent!")

		}
	//Note that we *must* wait for notifications, disconnecting immediately can cause the message we just wrote to be dropped
	//Of course it would be better if we could just quit upon receiving the notification
	fmt.Printf("Waiting for 5 seconds to get some notifications, if any.\n")
	}

	time.Sleep(1*time.Second)
	if !spam {
		close(done)
	}
	fmt.Println("Peripheral probe complete")
}

func padTo(s string, size int) ([]byte, int) {
	message := s 
	original_length := len(message)
	if original_length>size {
		original_length = size
	}
	message = pad.Right(message, size, "\x00")
	message = message[0:size]
	return []byte(message), original_length
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
	if os.Getenv("GODEBUG") == "" {
		os.Setenv("GODEBUG", "cgocheck=0")
	if err := syscall.Exec(os.Args[0], os.Args, os.Environ()); err != nil {
		log.Fatal(err)
	}
		/*cmd := exec.Command(os.Args[0], os.Args[1:]...)
		fmt.Printf(">%v\n",cmd)
		out, err := cmd.Output()
		fmt.Printf(">%s\n",out)
		fmt.Println(">",err)
*/
		
		
		fmt.Println("done")
		os.Exit(0)
	}
	fmt.Println(`

********************************************************************

If you can't find your watch, check to see if it is connected
to another program, and disconnect it there.  MacOSX automatically
takes control of every Bluetooth LE device near it

********************************************************************

`)
	peripheralIDs := flag.String("id", "", "ID of device to notify (get from --discover)")
	names := flag.String("name", "", "Send to every device with this name")
	flag.StringVar(&rawMessage, "text", "", "Message to send")
	discovers := flag.Bool("discover", false, "Scan for devices")
	flag.BoolVar(&spam, "spam", false, "Notify every matching device")
	flag.BoolVar(&verbose, "verbose", false, "Print extra information")
	flag.BoolVar(&verbose, "v", false, "Print extra information")
	flag.StringVar(&bandType, "type", "None", "Type of band, 'ID115' or 'HBand'")
	flag.Parse()
	discovery = *discovers
	if !discovery {
		if bandType == "None" {
			panic("You must choose a band type")
		}
	}
	//if *peripheralIDs=="" {
		//log.Fatalf("Peripheral ID must be given")
	//}
	peripheralID= strings.ToUpper(*peripheralIDs)
	name = *names
	if verbose {
		fmt.Println("Spam state: ", spam)
		fmt.Println("Sending message: |", rawMessage, "|")
	}

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
