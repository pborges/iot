package espiot

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const IdleTimeout = 3 * time.Second
const PingTimeout = 5 * time.Second

type Version struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

func parseVersion(v string) (Version, error) {
	matches := regexp.MustCompile("([0-9]+)\\.([0-9]+)").FindStringSubmatch(v)
	var ver Version
	var err error
	if len(matches) != 3 {
		err = fmt.Errorf("unexpected match count, got %d", len(matches))
		return ver, err
	}
	ver.Major, err = strconv.Atoi(matches[1])
	if err != nil {
		return ver, err
	}
	ver.Minor, err = strconv.Atoi(matches[2])
	if err != nil {
		return ver, err
	}
	return ver, nil
}

type DeviceInfo struct {
	Id               string   `json:"id"`
	Name             string   `json:"name"`
	Manufacturer     string   `json:"manufacturer"`
	Model            string   `json:"model"`
	HardwareVersion  Version  `json:"hardware_version"`
	FrameworkVersion Version  `json:"framework_version"`
	ControlAddress   net.Addr `json:"control_address"`
	UpdateAddress    net.Addr `json:"update_address"`
	ip               string
}

func (d DeviceInfo) IpAddr() string {
	return d.ip
}

func (d DeviceInfo) String() string {
	return fmt.Sprintf("id:%s model:%s hw:%s ver:%s name: %s",
		d.Id,
		d.Model,
		d.HardwareVersion,
		d.FrameworkVersion,
		d.Name,
	)
}

type Packet struct {
	Command  string
	Args     map[string]string
	response struct {
		Packets chan []Packet
		Error   chan error
	}
}

type Device struct {
	DeviceInfo
	Log *log.Logger

	attributes map[string]AttributeAndValue
	functions  map[string]Function

	lastWrite time.Time
	lastRead  time.Time
	connected bool

	control      chan Packet
	onConnect    func()
	onDisconnect func()
	onUpdate     func(AttributeAndValue)
}

func (d *Device) writeCommandAndReadResponse(conn net.Conn, p Packet) ([]Packet, error) {
	line := Encode(p)

	if line != "ping" {
		d.println("write:", line)
	}

	conn.SetDeadline(time.Now().Add(IdleTimeout))
	if _, err := fmt.Fprintln(conn, line); err != nil {
		return nil, err
	}

	scanr := bufio.NewScanner(conn)
	var res []Packet
	for scanr.Scan() {
		raw := scanr.Text()
		if line != "ping" {
			d.println("read:", raw)
		}
		if strings.HasPrefix(raw, "ok") {
			break
		} else {
			packet, err := Decode(raw)
			if err != nil {
				return nil, err
			}
			res = append(res, packet)
		}
	}
	if err := scanr.Err(); err != nil {
		d.println("err:", err)
		return nil, err
	}
	d.lastWrite = time.Now()
	return res, nil
}

func (d *Device) println(v ...interface{}) {
	if d.Log != nil {
		d.Log.Println(v)
	}
}

func (d *Device) printf(format string, v ...interface{}) {
	if d.Log != nil {
		d.Log.Printf(format, v)
	}
}

func (d *Device) OnConnect(fn func()) {
	if d.onConnect == nil {
		d.onConnect = fn
	} else {
		existing := d.onConnect
		d.onConnect = func() {
			existing()
			fn()
		}
	}
	if d.connected {
		fn()
	}
}

func (d *Device) OnDisconnect(fn func()) {
	if d.onDisconnect == nil {
		d.onDisconnect = fn
	} else {
		existing := d.onDisconnect
		d.onDisconnect = func() {
			existing()
			fn()
		}
	}
}

func (d *Device) OnUpdate(fn func(AttributeAndValue)) {
	if d.onUpdate == nil {
		d.onUpdate = fn
	} else {
		existing := d.onUpdate
		d.onUpdate = func(a AttributeAndValue) {
			existing(a)
			fn(a)
		}
	}
	if d.connected {
		for _, a := range d.attributes {
			fn(a)
		}
	}
}

func (d Device) Connected() bool {
	return d.connected
}

func (d *Device) Disconnect() error {
	return d.disconnect(errors.New("manual disconnect"))
}

func (d *Device) disconnect(err error) error {
	args := map[string]string{"err": ""}
	if err != nil {
		args["err"] = err.Error()
	}
	_, e := d.Exec(Packet{Command: "disconnect", Args: args})
	return e
}

func (d *Device) Reconnect() error {
	d.println("reconnect")
	return d.Connect(d.DeviceInfo.ControlAddress.String())
}

func (d *Device) Connect(addr string) error {
	d.ip = strings.Split(addr, ":")[0]

	if d.connected {
		// already connected, no problems
		return nil
	}
	var err error
	d.DeviceInfo.ControlAddress, err = net.ResolveTCPAddr("tcp", d.ip+":5000")
	if err != nil {
		return err
	}

	d.DeviceInfo.UpdateAddress, err = net.ResolveTCPAddr("tcp", d.ip+":5001")
	if err != nil {
		return err
	}

	d.println("dialing", d.DeviceInfo.ControlAddress.String())
	if conn, err := net.DialTimeout("tcp", d.DeviceInfo.ControlAddress.String(), IdleTimeout); err == nil {
		d.control = make(chan Packet)
		d.connected = true
		go func() {
			var err error
			for err == nil {
				select {
				case p := <-d.control:
					if p.Command == "disconnect" {
						err = fmt.Errorf("disconnect: %s", p.Args["err"])
						d.println(err)
						break
					}
					var res []Packet
					if res, err = d.writeCommandAndReadResponse(conn, p); err == nil {
						p.response.Packets <- res
					} else {
						p.response.Error <- err
					}
				case <-time.After(PingTimeout):
					//fmt.Println("ping")
					_, err = d.writeCommandAndReadResponse(conn, Packet{Command: "ping"})
				}
			}
			//fmt.Println("OnDisconnect", err)
			d.connected = false
			if d.onDisconnect != nil {
				d.println("fire onDisconnect")
				d.onDisconnect()
			}
		}()
		go func() {
			var disconnectErr error
			defer func() {
				_ = d.disconnect(disconnectErr)
			}()
			if conn, err := net.DialTimeout("tcp", d.DeviceInfo.UpdateAddress.String(), IdleTimeout); err == nil {
				scanr := bufio.NewScanner(conn)
				for scanr.Scan() {
					conn.SetDeadline(time.Now().Add(7 * time.Second))
					raw := scanr.Text()
					packet, err := Decode(raw)
					if err != nil {
						disconnectErr = fmt.Errorf("error: update decode %w", err)
						return
					}
					d.lastRead = time.Now()
					switch packet.Command {
					case "attr":
						d.println("update:", raw)
						if a, ok := d.attributes[packet.Args["name"]]; ok {
							if err = a.accept(packet.Args["value"]); err != nil {
								disconnectErr = fmt.Errorf("error: update parse int %s %w", packet.Args["value"], err)
								return
							} else if d.onUpdate != nil {
								d.onUpdate(a)
							}
						} else {
							disconnectErr = errors.New("error: update unknown attribute " + packet.Args["name"])
							return
						}
					}
				}
				if err := scanr.Err(); err != nil {
					disconnectErr = fmt.Errorf("error: update scan %w", err)
					return
				}
			}
		}()
	} else {
		//fmt.Println("OnDisconnectDial", err)
		d.connected = false
		if d.onDisconnect != nil {
			d.println("fire onDisconnect from failed dial")
			d.onDisconnect()
		}
		return err
	}

	if res, err := d.Exec(Packet{Command: "info"}); err == nil {
		if len(res) != 1 {
			return errors.New("unexpected response in info packet")
		}
		if id, ok := res[0].Args["id"]; ok {
			d.DeviceInfo.Id = id
		} else {
			return errors.New("no id in info packet")
		}
		if ver, ok := res[0].Args["ver"]; ok {
			d.DeviceInfo.FrameworkVersion, err = parseVersion(ver)
			if err != nil {
				return err
			}
		} else {
			return errors.New("no ver in info packet")
		}
		if hw, ok := res[0].Args["hw"]; ok {
			d.DeviceInfo.HardwareVersion, err = parseVersion(hw)
			if err != nil {
				return err
			}
		} else {
			return errors.New("no hw in info packet")
		}
		if m, ok := res[0].Args["model"]; ok {
			d.DeviceInfo.Model = m
		} else {
			return errors.New("no model in info packet")
		}
	} else {
		return err
	}

	if err := d.list(); err != nil {
		return err
	}

	//fmt.Println("OnConnect")
	if d.onConnect != nil {
		d.println("fire onConnect")
		d.onConnect()
	}

	return nil
}

func (d *Device) Exec(cmd Packet) ([]Packet, error) {
	if !d.connected {
		return nil, errors.New("not connected")
	}

	cmd.response.Packets = make(chan []Packet)
	cmd.response.Error = make(chan error)

	d.control <- cmd

	select {
	case res := <-cmd.response.Packets:
		return res, nil
	case err := <-cmd.response.Error:
		return nil, err
	case <-time.After(IdleTimeout):
		return nil, errors.New("timeout waiting for exec response")
	}
}

func (d Device) String() string {
	return d.DeviceInfo.String()
}

func (d *Device) Get(attr string) string {
	if a, ok := d.attributes[attr]; ok {
		return a.InspectValue()
	}
	return ""
}

func (d *Device) Set(attr string, value string) error {
	_, err := d.Exec(Packet{
		Command: "set",
		Args:    map[string]string{"name": attr, "value": value},
	})
	return err
}

func (d *Device) SetOnDisconnect(attr string, value string) error {
	_, err := d.Exec(Packet{
		Command: "set",
		Args:    map[string]string{"name": attr, "value": value, "disconnect": "true"},
	})
	return err
}

func (d *Device) GetBool(attr string) bool {
	if a, ok := d.attributes[attr]; ok {
		return a.(*BooleanAttributeValue).Value
	}
	return false
}

func (d *Device) SetBool(attr string, value bool) error {
	v := "false"
	if value {
		v = "true"
	}
	return d.Set(attr, v)
}

func (d *Device) SetBoolOnDisconnect(attr string, value bool) error {
	v := "false"
	if value {
		v = "true"
	}
	return d.SetOnDisconnect(attr, v)
}

func (d *Device) GetInteger(attr string) int {
	if a, ok := d.attributes[attr]; ok {
		return a.(*IntegerAttributeValue).Value
	}
	return 0
}

func (d *Device) SetInteger(attr string, value int) error {
	return d.Set(attr, strconv.Itoa(value))
}

func (d *Device) SetIntegerOnDisconnect(attr string, value int) error {
	return d.SetOnDisconnect(attr, strconv.Itoa(value))
}

func (d *Device) GetDouble(attr string) float64 {
	if a, ok := d.attributes[attr]; ok {
		return a.(*DoubleAttributeValue).Value
	}
	return 0
}

func (d *Device) SetDouble(attr string, value float64) error {
	return d.Set(attr, strconv.FormatFloat(value, 'E', 4, 64))
}

func (d *Device) SetDoubleOnDisconnect(attr string, value float64) error {
	return d.SetOnDisconnect(attr, strconv.FormatFloat(value, 'E', 4, 64))
}

func (d *Device) list() error {
	res, err := d.Exec(Packet{Command: "list"})
	if err != nil {
		return err
	}
	d.attributes = make(map[string]AttributeAndValue)
	d.functions = make(map[string]Function)
	for _, r := range res {
		if r.Command == "attr" {
			attr := Attribute{
				Name:      r.Args["name"],
				ReadOnly:  r.Args["readonly"] == "true",
				UpdatedAt: time.Now(),
			}
			switch r.Args["type"] {
			case "bool":
				d.attributes[r.Args["name"]] = &BooleanAttributeValue{
					Attribute: attr,
				}
			case "string":
				d.attributes[r.Args["name"]] = &StringAttributeValue{
					Attribute: attr,
				}
			case "integer":
				d.attributes[r.Args["name"]] = &IntegerAttributeValue{
					Attribute: attr,
				}
			case "double":
				d.attributes[r.Args["name"]] = &DoubleAttributeValue{
					Attribute: attr,
				}
			default:
				return errors.New("unknown attribute type: " + r.Args["type"])
			}
			d.attributes[r.Args["name"]].accept(r.Args["value"])
			if d.onUpdate != nil {
				d.onUpdate(d.attributes[r.Args["name"]])
			}
		} else if strings.HasPrefix(r.Command, "func") {
			if r.Command == "func" {
				d.functions[r.Args["name"]] = Function{
					Name: r.Args["name"],
				}
			} else if r.Command == "func.arg" {
				if fn, ok := d.functions[r.Args["func"]]; ok {
					fn.Args = append(fn.Args, FunctionArg{
						Name: r.Args["name"],
						Type: r.Args["type"],
					})
					d.functions[r.Args["func"]] = fn
				}
			}
		}
	}
	d.DeviceInfo.Name = d.attributes["config.name"].InspectValue()
	return nil
}

func (d *Device) ListAttributes() []AttributeAndValue {
	attrs := make([]AttributeAndValue, 0, len(d.attributes))
	for _, v := range d.attributes {
		attrs = append(attrs, v)
	}
	return attrs
}

func (d *Device) ListFunctions() []Function {
	fns := make([]Function, 0, len(d.functions))
	for _, v := range d.functions {
		fns = append(fns, v)
	}
	return fns
}
