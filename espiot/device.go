package espiot

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

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

type Packet struct {
	Command string
	Args    map[string]string
}

type Device struct {
	AlwaysReconnect bool
	Address         string
	Log             *log.Logger
	VerboseLogging  bool

	metadata struct {
		id    string
		model string
		hw    Version
		ver   Version
	}

	attributes map[string]AttributeAndValue
	functions  map[string]Function
	lock       sync.RWMutex

	connected    bool
	execute      chan request
	wg           sync.WaitGroup
	onConnect    func()
	onDisconnect func()
	onUpdate     func(*Device, AttributeAndValue)
	control      net.Conn
	update       net.Conn
}

type request struct {
	Packet
	Response chan []Packet
	Error    chan error
}

func (d Device) Id() string {
	return d.metadata.id
}

func (d Device) HardwareVersion() Version {
	return d.metadata.ver
}

func (d Device) FrameworkVersion() Version {
	return d.metadata.ver
}

func (d Device) Model() string {
	return d.metadata.model
}

func (d Device) Name() string {
	return d.GetString("config.name")
}

func (d *Device) Set(attr string, value interface{}) error {
	return d.set(attr, value, false)
}

func (d *Device) SetOnDisconnect(attr string, value interface{}) error {
	return d.set(attr, value, true)
}

func (d *Device) SetString(attr string, value string) error {
	return d.Set(attr, value)
}

func (d *Device) SetStringOnDisconnect(attr string, value string) error {
	return d.SetOnDisconnect(attr, value)
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

func (d *Device) SetInteger(attr string, value int) error {
	return d.Set(attr, strconv.Itoa(value))
}

func (d *Device) SetIntegerOnDisconnect(attr string, value int) error {
	return d.SetOnDisconnect(attr, strconv.Itoa(value))
}

func (d *Device) SetDouble(attr string, value float64) error {
	return d.Set(attr, strconv.FormatFloat(value, 'f', 4, 64))
}

func (d *Device) SetDoubleOnDisconnect(attr string, value float64) error {
	return d.SetOnDisconnect(attr, strconv.FormatFloat(value, 'f', 4, 64))
}

func (d *Device) Get(attr string) interface{} {
	d.lock.RLock()
	defer d.lock.RUnlock()
	if a, ok := d.attributes[attr]; ok {
		return a.Interface()
	}
	return nil
}
func (d *Device) GetString(attr string) string {
	d.lock.RLock()
	defer d.lock.RUnlock()
	if a, ok := d.attributes[attr]; ok {
		return a.(*StringAttributeValue).Value
	}
	return ""
}
func (d *Device) GetBool(attr string) bool {
	d.lock.RLock()
	defer d.lock.RUnlock()
	if a, ok := d.attributes[attr]; ok {
		return a.(*BooleanAttributeValue).Value
	}
	return false
}

func (d *Device) GetInteger(attr string) int {
	d.lock.RLock()
	defer d.lock.RUnlock()
	if a, ok := d.attributes[attr]; ok {
		return a.(*IntegerAttributeValue).Value
	}
	return 0
}
func (d *Device) GetDouble(attr string) float64 {
	d.lock.RLock()
	defer d.lock.RUnlock()
	if a, ok := d.attributes[attr]; ok {
		return a.(*DoubleAttributeValue).Value
	}
	return 0
}

func (d *Device) Exec(req Packet) ([]Packet, error) {
	if !d.connected {
		return nil, errors.New("not connected")
	}
	request := request{
		Packet:   req,
		Response: make(chan []Packet),
		Error:    make(chan error),
	}
	if req.Command != "ping" {
		d.log(false).Println("exec:", Encode(req))
	} else {
		d.log(true).Println("exec:", Encode(req))
	}
	select {
	case d.execute <- request:
	case <-time.After(5 * time.Second):
		return nil, errors.New("exec request timeout")
	}
	select {
	case res := <-request.Response:
		return res, nil
	case err := <-request.Error:
		if req.Command != "ping" {
			d.log(false).Println("error exec:", Encode(req), "error:", err)
		} else {
			d.log(true).Println("error exec:", Encode(req), "error:", err)
		}
		return nil, err
	}
}

func (d *Device) Connect() error {
	if d.connected {
		// already connected
		return nil
	}
	if d.Address == "" {
		return errors.New("address cannot be nil")
	}
	if d.connected {
		return errors.New("already connected")
	}
	d.wg.Wait()
	for {
		if err := d.dial(); err == nil {
			break
		} else if d.AlwaysReconnect {
			d.log(true).Println("attempt dial error:", err)
		} else {
			return err
		}
	}

	d.wg.Add(1)
	go d.handleUpdates()

	d.wg.Add(1)
	d.execute = make(chan request)
	go d.handleControl()

	d.connected = true
	if err := d.init(); err != nil {
		return err
	}

	if d.onConnect != nil {
		d.log(false).Println("connect")
		go d.onConnect()
	}
	go func() {
		d.wg.Wait()
		d.connected = false
		if d.onDisconnect != nil {
			d.log(false).Println("disconnect")
			go d.onDisconnect()
		}
		if d.AlwaysReconnect {
			d.log(false).Println("reconnecting...")
			if err := d.Connect(); err != nil {
				d.log(true).Println("reconnect failed", err)
			}
		}
	}()
	return nil
}

func (d *Device) Disconnect() {
	alwaysReconnect := d.AlwaysReconnect
	d.AlwaysReconnect = false
	d.disconnect()
	d.wg.Wait()
	d.AlwaysReconnect = alwaysReconnect
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
		go fn()
	}
}

func (d *Device) OnDisconnect(fn func()) {
	if d.onDisconnect == nil {
		d.onDisconnect = fn
	} else {
		existing := d.onDisconnect
		d.onDisconnect = func() {
			existing()
			go fn()
		}
	}
}

func (d *Device) OnUpdate(fn func(*Device, AttributeAndValue)) {
	if d.onUpdate == nil {
		d.onUpdate = fn
	} else {
		existing := d.onUpdate
		d.onUpdate = func(d *Device, a AttributeAndValue) {
			existing(d, a)
			fn(d, a)
		}
	}
	if d.connected {
		d.lock.RLock()
		for _, a := range d.attributes {
			go fn(d, a)
		}
		d.lock.RUnlock()
	}
}

func (d *Device) ListAttributes() []AttributeAndValue {
	d.lock.RLock()
	defer d.lock.RUnlock()
	attrs := make([]AttributeAndValue, 0, len(d.attributes))
	for _, v := range d.attributes {
		attrs = append(attrs, v)
	}
	return attrs
}

func (d *Device) ListFunctions() []Function {
	d.lock.RLock()
	defer d.lock.RUnlock()
	fns := make([]Function, 0, len(d.functions))
	for _, v := range d.functions {
		fns = append(fns, v)
	}
	return fns
}

func (d *Device) init() error {
	res, err := d.Exec(Packet{Command: "info"})
	if err != nil {
		return err
	}

	if err := d.setMetadata(res[0]); err != nil {
		return err
	}

	res, err = d.Exec(Packet{Command: "list"})
	if err != nil {
		return err
	}

	return d.handleList(res)
}

func (d *Device) handleList(res []Packet) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.attributes == nil {
		d.attributes = make(map[string]AttributeAndValue)
	}
	if d.functions == nil {
		d.functions = make(map[string]Function)
	}
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
			if err := d.attributes[r.Args["name"]].accept(r.Args["value"]); err != nil {
				return err
			}
			if d.onUpdate != nil {
				d.onUpdate(d, d.attributes[r.Args["name"]])
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
	return nil
}

func (d *Device) setMetadata(p Packet) (err error) {
	if p.Command != "info" {
		err = errors.New("invalid packet")
		return
	}

	if v, ok := p.Args["id"]; ok && v != "" {
		d.metadata.id = v
	} else {
		err = errors.New("id missing from info packet")
		return
	}

	if v, ok := p.Args["model"]; ok && v != "" {
		d.metadata.model = v
	} else {
		err = errors.New("model missing from info packet")
		return
	}

	if v, ok := p.Args["hw"]; ok && v != "" {
		d.metadata.hw, err = parseVersion(v)
		if err != nil {
			return
		}
	} else {
		err = errors.New("hardware version missing from info packet")
		return
	}

	if v, ok := p.Args["ver"]; ok && v != "" {
		d.metadata.ver, err = parseVersion(v)
		if err != nil {
			return
		}
	} else {
		err = errors.New("framework version missing from info packet")
		return
	}

	return
}

func (d *Device) set(attr string, value interface{}, disconnect bool) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	if _, ok := d.attributes[attr]; !ok {
		return errors.New("unknown attribute")
	}

	req := Packet{
		Command: "set",
		Args:    make(map[string]string),
	}
	if disconnect {
		req.Args["disconnect"] = "true"
	}
	req.Args["name"] = attr
	switch value.(type) {
	case int, int8, int16, int32, int64:
		req.Args["value"] = fmt.Sprintf("%d", value)
	case float32, float64:
		req.Args["value"] = fmt.Sprintf("%f", value)
	case bool:
		req.Args["value"] = fmt.Sprintf("%t", value)
	case string:
		req.Args["value"] = value.(string)
	default:
		return errors.New("unknown data type")
	}
	_, err := d.Exec(req)
	return err
}

func (d Device) log(verbose bool) *log.Logger {
	if d.Log == nil || verbose && !d.VerboseLogging {
		return log.New(ioutil.Discard, "", 0)
	}
	return d.Log
}

func (d *Device) exec(conn net.Conn, scanner *bufio.Scanner, req request) error {
	writeTimeout := 2 * time.Second
	readTimeout := 2 * time.Second

	if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		err = fmt.Errorf("set write deadline: %w", err)
		req.Error <- err
		return err
	}
	if _, err := fmt.Fprintln(conn, Encode(req.Packet)); err != nil {
		err = fmt.Errorf("encode: %w", err)
		req.Error <- err
		return err
	}

	if err := conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		err = fmt.Errorf("set read deadline: %w", err)
		req.Error <- err
		return err
	}
	var res []Packet
	for scanner.Scan() {
		cmd, err := Decode(scanner.Text())
		if err != nil {
			err = fmt.Errorf("decode: %w", err)
			req.Error <- err
			return err
		}
		if cmd.Command != "ok" {
			res = append(res, cmd)
		} else {
			req.Response <- res
			return nil
		}
		if err := conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
			err = fmt.Errorf("set read deadline: %w", err)
			req.Error <- err
			return err
		}
	}
	req.Error <- scanner.Err()
	return scanner.Err()
}

func (d *Device) handleControl() {
	d.log(true).Println("[control   ] open")
	var err error
	defer func() {
		d.log(true).Printf("[control   ] close error: %s\n", err)
		d.wg.Done()
		d.disconnect()
	}()

	scanner := bufio.NewScanner(d.control)
	for {
		select {
		case req := <-d.execute:
			d.log(true).Println("[control   ] write", Encode(req.Packet))
			if err = d.exec(d.control, scanner, req); err != nil {
				return
			}
		case <-time.After(5 * time.Second):
			req := request{
				Packet:   Packet{Command: "ping"},
				Response: make(chan []Packet, 1),
				Error:    make(chan error, 1),
			}
			d.log(true).Println("[control   ] ping")
			if err = d.exec(d.control, scanner, req); err != nil {
				return
			}
			select {
			case <-req.Response:
			case err = <-req.Error:
				return
			}
		}
	}
}

func (d *Device) handleUpdates() {

	readTimeout := 7 * time.Second
	d.log(true).Println("[update    ] open")
	var err error
	defer func() {
		d.log(true).Printf("[update    ] close error: %s\n", err)
		d.wg.Done()
		d.disconnect()
	}()

	scanner := bufio.NewScanner(d.update)
	if err = d.update.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		return
	}
	for scanner.Scan() {
		d.log(true).Printf("[update    ] read: %s\n", scanner.Text())
		var packet Packet
		packet, err = Decode(scanner.Text())
		if err != nil {
			return
		}

		if packet.Command != "ping" {
			d.log(false).Println("update:", Encode(packet))
		} else {
			d.log(true).Println("update:", Encode(packet))
		}

		switch packet.Command {
		case "attr":
			err = d.handleUpdate(packet.Args["name"], packet.Args["value"])
			if err != nil {
				return
			}
		}

		if err = d.update.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
			return
		}
	}
	err = scanner.Err()
}

func (d *Device) handleUpdate(name string, value string) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	if a, ok := d.attributes[name]; ok {
		if err := a.accept(value); err != nil {
			return fmt.Errorf("error: update parse int %s %w", value, err)
		} else if d.onUpdate != nil {
			go d.onUpdate(d, a)
		}
	} else {
		return errors.New("error: update unknown attribute " + name)
	}
	return nil
}

func (d *Device) dial() (err error) {
	d.control, err = net.DialTimeout("tcp", d.Address+":5000", 3*time.Second)
	if err != nil {
		return fmt.Errorf("unable to dial control %w", err)
	}

	d.update, err = net.DialTimeout("tcp", d.Address+":5001", 3*time.Second)
	if err != nil {
		d.disconnect()
		return fmt.Errorf("unable to dial update %w", err)
	}
	return nil
}

func (d *Device) disconnect() {
	if d.control != nil {
		d.control.Close()
	}
	if d.update != nil {
		d.update.Close()
	}
}
