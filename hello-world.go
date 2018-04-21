package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
)

const (
	//Acceleration of gravety in m/(s*s)
	g = 9.81
	//for conversion from gain to kg
	kg_in_grain = 0.0000648
	// foot to meter
	ft_to_meter = 0.3048
)

var (
	// time delta in seconds
	dt float64
	//mass in grains
	//m_grain float64
	//mass in kilograms
	m float64
	// Drag
	D float64
	//velocity
	v float64
	//angle from horizon in radians
	theta float64
	//caliber in mm or inches
	//caliber float64
	//units for caliber ,possible values 's' for standard or 'm' metric
	//caliber_unit string
	//units for muzzle velocity ,possible values 's' for standard or 'm' metric
	//velocity_unit string
	//air density
	//rho float64
	//drag coefecient
	//C float64
	//bullet frontal area meters squared
	//A float64
	//find radius in meters
	r float64

	//delta map[string]float64

	//motion    map[string]float64
	//delta_map map[string]float64
	//position  map[string]float64
)

type motion struct {
	x     float64
	y     float64
	vSubx float64
	vSuby float64
	aSubx float64
	aSuby float64
	time  float64
}

type delta struct {
	deltavx float64
	deltavy float64
	deltax  float64
	deltay  float64
}

type variables struct {
	m    float64
	v    float64
	A    float64
	bc   float64
	rho  float64
	step float64
}

type input_stuff struct {
	Mass         int    `json:"mass"`
	Velocity     int    `json:"velocity"`
	Velocityunit string `json:"velocityunit"`
	Caliber      int    `json:"caliber"`
	Caliberunit  string `json:"caliberunit"`
}

func grain_to_kg(m_grain float64) float64 {
	m = m_grain * kg_in_grain
	return m
}

func fps_to_mps(velocity_unit string, v float64) (float64, error) {
	if velocity_unit == "m" {
		return v, nil
	} else if velocity_unit == "s" {
		return v * ft_to_meter, nil
	} else {
		return -1, errors.New("Non-supported velocity unit, choose m or s")
	}
}

func area(caliber float64, unit string) (float64, error) {
	if unit == "s" {
		const metersInInch = .0254
		r = caliber * metersInInch / 2
	} else if unit == "m" {
		r = caliber / 2000
	} else {
		return -1, errors.New("Non-supported caliber unit, choose m or s")
	}
	return math.Pi * math.Pow(r, 2), nil
}

func drag(rho float64, C float64, A float64) float64 {
	//calcuate drag based on pit velocity and rho
	D := rho * C * A / 2
	return D
}

func calculate_rho() float64 {
	//	var T float32            //temperature in degrees C
	//	var P float32            //Pressure in in hPa
	//	var D float32            // dew point in degrees C
	var rho float64 = 1.29 //currently set to static, to be compleated later
	return rho
}

func cd_from_bc(bc float64, m float64, d float64) float64 {
	const cg_g7 = 0.195
	const cg_g1 = 0.36 //0.51
	i := (40.00000000 / 7000) / (math.Pow(0.22, 2) * bc)
	cp := cg_g1 * i
	fmt.Println("cp ", cp)
	return cp
}

func initial_motion_components(v float64, theta float64) *motion {
	//point in time velocity in x direction
	vSuby := v * math.Sin(theta)
	//point in time velocity in y direction
	vSubx := v * math.Cos(theta)
	m := &motion{}
	m.vSubx = vSubx
	m.vSuby = vSuby
	return m
}

func acceleration_components(m float64, D float64, mo *motion, dt *float64) {
	//point in time acceleration in x direction
	aSuby := -9.806 + (D/m)*math.Pow(mo.vSuby, 1)
	//point in time acceleration in y direction
	aSubx := (-D / m) * math.Pow(mo.vSubx, 2)
	//map of point in time motion components
	mo.aSubx = aSubx
	mo.aSuby = aSuby
	mo.time += *dt
}

func find_delta(mo *motion, dt *float64, d *delta) {
	deltax := (mo.vSubx * *dt) + 0.5*(mo.aSubx*math.Pow(*dt, 2))
	deltay := (mo.vSuby * *dt) + 0.5*(mo.aSuby*math.Pow(*dt, 2))
	deltavx := mo.aSubx * *dt
	deltavy := mo.aSuby * *dt
	d.deltax = deltax
	d.deltay = deltay
	d.deltavx = deltavx
	d.deltavy = deltavy
}

func new_components(mo *motion, d *delta) {
	mo.vSubx += d.deltavx
	mo.vSuby += d.deltavy
	mo.x += d.deltax
	mo.y += d.deltay
}

func find_zero(res map[float64]map[string]float64, zero float64) (float64, error) {
	for i := 0.0001; i < .020; i += .0001 {
		fmt.Println(res[i])
		if res[i]["range"] >= zero {
			zero_theta := i
			fmt.Println(zero_theta)
			return zero_theta, nil
		}
	}
	return -1, errors.New("please use a valid zero distance")
}

func compute_variables() *variables {
	vars := &variables{}
	vars.m = grain_to_kg(40)
	vars.v, _ = fps_to_mps("s", 1090)
	vars.A, _ = area(0.224, "s")
	vars.rho = calculate_rho()
	vars.bc = 0.122
	return vars
}

func read_config(filename string) *input_stuff {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Printf("%s\n", raw)

	var in input_stuff
	json.Unmarshal(raw, &in)
	fmt.Printf("%+v", in)
	return &in
}

func main() {
	vars := compute_variables()
	fmt.Printf("%+v", vars)
	cd := cd_from_bc(vars.bc, vars.m, 0.00569)
	D := drag(vars.rho, cd, vars.A)
	dt := .0001
	d := delta{}
	res := make(map[float64]map[string]float64)
	mo := initial_motion_components(vars.v, 0)
	for i := 0.0001; i < .020; i += .0001 {
		mo = initial_motion_components(vars.v, i)
		for {
			if mo.y >= 0.0 {
				acceleration_components(vars.m, D, mo, &dt)
				find_delta(mo, &dt, &d)
				new_components(mo, &d)
			} else {
				acceleration_components(vars.m, D, mo, &dt)
				find_delta(mo, &dt, &d)
				new_components(mo, &d)
				res[i] = map[string]float64{
					"range": mo.x,
					"time":  mo.time}
				//fmt.Printf("%.4f", i)
				//fmt.Println(" ", res[i])
				break
			}
		}
	}
	zero, _ := find_zero(res, 50)
	fmt.Println(zero)
	for i := 0.0001; i < .020; i += .0001 {
		fmt.Print("dope: ")
		fmt.Printf("%.4f", i-zero)
		fmt.Print(" range: ")
		fmt.Printf("%.2f", res[i]["range"])
		fmt.Print(" time: ")
		fmt.Printf("%.2f", res[i]["time"])
		fmt.Println()
	}
	in := read_config("22lr.json") //fmfmt.Printf("%.4f"t.Printf("%+v\n", mo)
	fmt.Println(in)
}
