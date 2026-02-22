package main
import "log"
import "fmt"
import "os"
import "time"
import "github.com/sstallion/go-hid"

func main() {
	log.SetPrefix("vaxee-read-battery: ")
	path := ""

	if err := hid.Init(); err != nil {
		fmt.Println("-1")
		log.Fatal(err)
	}

	hid.Enumerate(0x3057, hid.ProductIDAny, func(info *hid.DeviceInfo) error {
		if info.Usage == 0x01  && info.UsagePage == 0xff05 {
			fmt.Fprintf(os.Stderr, "ID %04x:%04x %s %s (0x%x 0x%x): %s\n",
				info.VendorID,
				info.ProductID,
				info.MfrStr,
				info.ProductStr,
				info.Usage,
				info.UsagePage,
				info.Path,
				)
			path = info.Path
		}
		return nil
	})

	if path == "" {
		fmt.Println("-1")
		log.Fatal("no mouse device found")
	}

	d, err := hid.OpenPath(path)
	if err != nil {
		fmt.Println("-1")
		log.Fatal(err)
	}

	b := make([]byte, 64)
	b[0] = 0x0e  // report id
	b[1] = 0xa5  // header
	b[2] = 0x0b  // command id
	b[3] = 0x01  // read (0x01) or write (0x02)
	b[4] = 0x01  // ret len
	if _, err := d.SendFeatureReport(b); err != nil {
		fmt.Println("-1")
		log.Println("SendFeatureReport: ", err)
	}

	tries := 0
	for tries < 10 {
		tries += 1
		if _, err := d.GetFeatureReport(b); err != nil {
			fmt.Println("-1")
			log.Fatal("GetFeatureReport: ", err)
		}
		if b[5] > 0 && b[5] < 20 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// happens for ~500ms after mouse turning on
	if b[5] == 255 {
		fmt.Println("-1")
		log.Fatal("mouse returned garbage data")
	}

	// number is in 5% increments
	fmt.Printf("%d\n", b[5]*5)

	if err := hid.Exit(); err != nil {
		log.Fatal(err)
	}
}
