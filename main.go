package main


import (
	"os"
	"os/exec"
	"fmt"
	"flag"
	"strings"
	"errors"
        //"encoding/json"
	//"math"
       "github.com/seldonsmule/simpleconffile"
       "github.com/seldonsmule/smartthings"
       "github.com/seldonsmule/powerwall"
       "github.com/seldonsmule/logmsg"
        "time"
        //"github.com/seldonsmule/restapi"

)


type Configuration struct {

  ST_Token string             // Token from SmartThings to access your devices
  ConfFilename string         // Name of conffile
  Encrypted bool

  PW_Userid string
  PW_Passwd string

  ST_DownScene string
  ST_UpScene string

  DownScript string


/*
  // non users input

  ST_Scenes smartthings.StScenes // holds the data from a screns querry to ST
  ST_Devices smartthings.StDevices // holds the data from a screns querry to ST
*/

}


const COMPILE_IN_KEY = "example key 9999"

const CERTFILE string = "powerwall.cer"

var gMyConf Configuration
var gsKey string = "notset"


func help(){

  fmt.Println("Test if grid is down and execute smartthing scene")

  fmt.Println("Usage griddown -cmd [a command, see below] [-rundir runpath] [-key encryptkey for config file]")
  fmt.Println()
  flag.PrintDefaults()
  fmt.Println()
  fmt.Println("cmds:")
  fmt.Println("       setconf - Setup Conf file")
  fmt.Println("             -sttoken SmartThings API token")
  fmt.Println("             -stdownscene SmartThings Scene to call with grid is down")
  fmt.Println("             -stupcene SmartThings Scene to call with grid is up")
  fmt.Println("             -pwuserid Powerwall user id")
  fmt.Println("             -pwpasswd Powerwall password")
  fmt.Println("             -downscript scriptname to run when grid is down")
  fmt.Println("             -conffile name of conffile (.griddown.conf default)")
  fmt.Println("       readconf - Display conf info")
  fmt.Println("       convertkey - Convert config file from passed in key to internally complied/generated key")
  fmt.Println("       gridstatus - Test if Power Grid is up or down")
  fmt.Println("       rundownscript - Run a script when grid is down")
  fmt.Println("       runscene - Runs a SmartThings Scene")
  fmt.Println("             -name Name of a predefined SmartThings scene")
  fmt.Println("       listscenes - List all the SmartThings Scense that are avaialble")
  fmt.Println("       switchon - Turns on a switch")
  fmt.Println("             -name Name of device (switch)")
  fmt.Println("       switchon - Turns off a switch")
  fmt.Println("             -name Name of device (switch)")
  fmt.Println("       switchstatus - status of a switch state")
  fmt.Println("             -name Name of device (switch)")
  fmt.Println()


}

func rundownscript() bool{

  fmt.Println("Running down script: ", gMyConf.DownScript)

  cmd := exec.Command(gMyConf.DownScript)

  fmt.Println(cmd)

  err := cmd.Run()

  if(err != nil){
    fmt.Println("Error running down script: ", err)
    return false
  }

  return true
}

func saveconf() bool {

  simple := simpleconffile.New(getencryptkey(), gMyConf.ConfFilename)

  if(gMyConf.Encrypted){
    gMyConf.ST_Token = simple.EncryptString(gMyConf.ST_Token)
    gMyConf.PW_Userid = simple.EncryptString(gMyConf.PW_Userid)
    gMyConf.PW_Passwd = simple.EncryptString(gMyConf.PW_Passwd)
  }

  simple.SaveConf(gMyConf)

  return true
}

func readconf(confFile string, printstd bool) bool{

  simple := simpleconffile.New(getencryptkey(), confFile)

  if(!simple.ReadConf(&gMyConf)){
    msg := fmt.Sprintln("Error reading conf file: ", confFile)
    logmsg.Print(logmsg.Warning, msg)
    return false
  }

  if(gMyConf.Encrypted){
    gMyConf.ST_Token = simple.DecryptString(gMyConf.ST_Token)
    gMyConf.PW_Userid = simple.DecryptString(gMyConf.PW_Userid)
    gMyConf.PW_Passwd = simple.DecryptString(gMyConf.PW_Passwd)
  }

     
  if(printstd){

    fmt.Printf("Encrypted [%v]\n", gMyConf.Encrypted)
    fmt.Printf("ST_Token [%v]\n", gMyConf.ST_Token)
    fmt.Printf("ST_DownScene [%v]\n", gMyConf.ST_DownScene)
    fmt.Printf("ST_UpScene [%v]\n", gMyConf.ST_UpScene)
    fmt.Printf("ConfFilename [%v]\n", gMyConf.ConfFilename)
    fmt.Printf("DownScript [%v]\n", gMyConf.DownScript)
    fmt.Printf("PW_Userid [%v]\n", gMyConf.PW_Userid)
    fmt.Printf("PW_Passwd [%v]\n", gMyConf.PW_Passwd)

  }

  return true

}

func testLockfile() bool {

  lockfile := fmt.Sprintf("%s/tmp/griddown.lck", os.Getenv("HOME"))

  _, statErr := os.Stat(lockfile)

  if(os.IsNotExist(statErr)){
    return false
  }

  return true;

}

func deleteLockfile(){

  lockfile := fmt.Sprintf("%s/tmp/griddown.lck", os.Getenv("HOME"))

  _, statErr := os.Stat(lockfile)

  // if lock file already exist - just log it and exit
  if(statErr == nil){

    //fmt.Println("Lockfile ", lockfile, " created: ",info.ModTime())
    os.Remove(lockfile);

    return;

  }else{
    fmt.Println("Lockfile already deleted");
  }

}

func createLockfile(){

  lockfile := fmt.Sprintf("%s/tmp/griddown.lck", os.Getenv("HOME"))

  info, statErr := os.Stat(lockfile)

  // if lock file already exist - just log it and exit
  if(statErr == nil){

    fmt.Println("Lockfile ", lockfile, " created: ",info.ModTime())

    return;

  }

  // otherwise create it

  lockWriteFile, openErr := os.Create(lockfile)

  if(openErr != nil){

    fmt.Println("Error creating lockfile: ", lockfile );

    return;

  }

  fmt.Println("Created lockfile: ", lockfile );

  lockWriteFile.Close()


}


func gridstatus(pw *powerwall.Powerwall, st *smartthings.SmartThings) error{

  err, status := gridup(pw)

  if(err != nil){
    logmsg.Print(logmsg.Error, err)
    return(err)
  }

  if(status){
    fmt.Println("Grid is working")
    logmsg.Print(logmsg.Info, "Grid is working")

    // 1st see if we have already done something about this

    if(testLockfile()){ // if true - grid used to be down

      fmt.Printf("Running SmartThings GridUp Scene: %s\n", gMyConf.ST_UpScene)
      st.RunScene(gMyConf.ST_UpScene)
      
      deleteLockfile()

    }
        

  }else{
    fmt.Println("Yikes - power is down")
    logmsg.Print(logmsg.Info, "Yikes - power is down")

    if(!testLockfile()){ // if false - grid used to be up

      fmt.Printf("Running SmartThings GridDown Scene: %s\n", gMyConf.ST_DownScene)
      st.RunScene(gMyConf.ST_DownScene)

      rundownscript()
      
      createLockfile()

    }

  }

  return nil
}

func gridup(pw *powerwall.Powerwall) (error, bool){

  if(!pw.Login(gMyConf.PW_Userid, gMyConf.PW_Passwd, false)){
    err := errors.New("Powerwall login failed")
    logmsg.Print(logmsg.Error, err)
    return err, false
  }

  worked, ss := pw.GetSystemStatus()

  if(!worked){
    err := errors.New("Powerwall GetSystemStatus() failed")
    logmsg.Print(logmsg.Error, err)
    return err, false
  }

  msg := fmt.Sprintf("System Status[%s]\n", ss.SystemIslandState)

  logmsg.Print(logmsg.Info, msg)

  if(ss.SystemIslandState == "SystemGridConnected"){

    logmsg.Print(logmsg.Info,"We have power from the grid!")
    return nil, true

  }

  logmsg.Print(logmsg.Info,"RUNNING ON SOLAR BACKUP")

  return nil, false
}

// cheezy - but works - we need a 16 character string
// WARNING - this is for MacOs - you will need (if you want to use this)
//           to  update for your OS logic
//
func getencryptkey() string{

  var serialNumber string
  size := 16
  var key string

  if(gsKey != "notset"){ // someone passed it in
    key = gsKey
  }else{

    out, _ := exec.Command("/usr/sbin/ioreg", "-l").Output() // err ignored for brevity

    for _, l := range strings.Split(string(out), "\n") {
      if strings.Contains(l, "IOPlatformSerialNumber") {
        s := strings.Split(l, " ")
        serialNumber = s[len(s)-1]
        break
      }
    }

    for _, e := range strings.Split(string(serialNumber), "\"") {

      if(len(e) > 0){
        serialNumber = e
        break
      }
  
    }

    //serialNumber = serialNumber + "AAAA"

    length := len(serialNumber)

    //fmt.Printf("serial is [%s] with length[%d]\n", serialNumber, length)


    if(length == size){
      key = serialNumber
    }else if(length < size){

      padlen := size-length

      pad := "9"

      for i := 1; i < padlen; i++ {

        pad = pad + "9"

      }

      key = serialNumber + pad

    }else{
     // fmt.Printf("crap greater than [%d]\n" ,size)

      key = serialNumber[: + size]
    }

  }

  //fmt.Println(key)

  //return(COMPILE_IN_KEY)
 
  return(key)
}


func main() {

  cmdPtr := flag.String("cmd", "help", "Command to run")
  rundirPtr := flag.String("rundir", "./", "Path config file and write log to")
  keyPtr := flag.String("key", "notset", "Pass-in instead of use a complied in encrypt key for the config file")
  sttokenPtr := flag.String("sttoken", "notset", "SmartThings access Token")
  stdownscenePtr := flag.String("stdownscene", "notset", "SmartThings Scene for when grid is down]")
  stupscenePtr := flag.String("stupscene", "notset", "SmartThings Scene for when grid is up]")
  namePtr := flag.String("name", "notset", "SmartThings Device/Scene to call - used with cmd [switchon | switchoff]")
  confPtr := flag.String("conffile", ".griddown.conf", "config file name")
  bdebugPtr := flag.Bool("debug", false, "If true, do debug magic")

  pwuserPtr := flag.String("pwuser", "notset", "Powerwall Userid")
  pwpasswdPtr := flag.String("pwpasswd", "notset", "Powerwall Password")
  downscriptPtr := flag.String("downscript", "notset", "Script to run when grid goes down")



  flag.Parse()

  fmt.Printf("cmd=%s rundir=%s\n", *cmdPtr, *rundirPtr)

  cderr := os.Chdir(*rundirPtr)

  gsKey = *keyPtr

  if(cderr != nil){
    msg := fmt.Sprintf("Error with chdir: %s", cderr)
    logmsg.Print(logmsg.Error,msg)
    fmt.Println(msg)
    os.Exit(2)
  }

  logmsg.SetLogFile("griddown.log");

  logmsg.Print(logmsg.Info, "cmdPtr = ", *cmdPtr)
  logmsg.Print(logmsg.Info, "rundirPtr = ", *rundirPtr)
  logmsg.Print(logmsg.Info, "keyPtr = ", *keyPtr)
  logmsg.Print(logmsg.Info, "confPtr = ", *confPtr)
  logmsg.Print(logmsg.Info, "sttokenPtr = ", *sttokenPtr)
  logmsg.Print(logmsg.Info, "stdownscenePtr = ", *stdownscenePtr)
  logmsg.Print(logmsg.Info, "stupscenePtr = ", *stupscenePtr)
  logmsg.Print(logmsg.Info, "namePtr = ", *namePtr)
  logmsg.Print(logmsg.Info, "pwuserPtr = ", *pwuserPtr)
  logmsg.Print(logmsg.Info, "pwpasswdPtr = ", *pwpasswdPtr)
  logmsg.Print(logmsg.Info, "bdebugPtr = ", *bdebugPtr)
  logmsg.Print(logmsg.Info, "downscriptPtr = ", *downscriptPtr)
  logmsg.Print(logmsg.Info, "tail = ", flag.Args())

  if(*cmdPtr == "help"){
    help()
    os.Exit(1)
  }


  readconf(*confPtr, false)

  st := smartthings.New()

  st.SetToken(gMyConf.ST_Token)

  pw := powerwall.New(CERTFILE)

  //st.Dump()

  switch *cmdPtr {

    case "rundownscript":
      if(rundownscript() == false){
	fmt.Println("Error running script")
	fmt.Println("%s", gMyConf.DownScript)
	os.Exit(1)
      }


    case "gridstatus":

      err := gridstatus(pw, st)

      if(err != nil){
        fmt.Println(err)
        os.Exit(1)
      }


    case "login":

      if(!pw.Login(gMyConf.PW_Userid, gMyConf.PW_Passwd, false)){
        fmt.Printf("Powerwall login faield\n")
        os.Exit(2)
      }

      gridup(pw)

    case "readconf":
      fmt.Println("Reading conf file")
      readconf(*confPtr, true)

    case "convertkey":

      if(gsKey == "notset"){
        fmt.Println("You must pass in the old encrypt key (-key) to convert")
        os.Exit(1)
      }

      // reading with passed in key
      readconf(*confPtr, true)


      gsKey = "notset"

      saveconf()

    case "setconf":

      readconf(*confPtr, false) // ignore errors

      fmt.Println("Setting conf file")

      gMyConf.Encrypted = true

      if(strings.Compare(*sttokenPtr, "notset") != 0){
        gMyConf.ST_Token = *sttokenPtr
      }else{
        gMyConf.ST_Token = gMyConf.ST_Token
      }

      if(strings.Compare(*downscriptPtr, "notset") != 0){
        gMyConf.DownScript = *downscriptPtr
      }else{
        gMyConf.DownScript = gMyConf.DownScript
      }

      if(strings.Compare(*pwuserPtr, "notset") != 0){
        gMyConf.PW_Userid = *pwuserPtr
      }else{
        gMyConf.PW_Userid = gMyConf.PW_Userid
      }

      if(strings.Compare(*pwpasswdPtr, "notset") != 0){
        gMyConf.PW_Passwd = *pwpasswdPtr
      }else{
        gMyConf.PW_Passwd = gMyConf.PW_Passwd
      }


      gMyConf.ConfFilename = *confPtr

      if(strings.Compare(*stdownscenePtr, "notset") != 0){

        if(!st.ValidateScene(*stdownscenePtr)){
          fmt.Printf("Invalid stdownscene[%s] - Run listscenes\n", *stdownscenePtr)
          os.Exit(1)
        }else{
          gMyConf.ST_DownScene = *stdownscenePtr
        }

      }

      if(strings.Compare(*stupscenePtr, "notset") != 0){

        if(!st.ValidateScene(*stupscenePtr)){
          fmt.Printf("Invalid stupscene[%s] - Run listscenes\n", *stupscenePtr)
          os.Exit(1)
        }else{
          gMyConf.ST_UpScene = *stupscenePtr
        }
      }

      saveconf()

      readconf(*confPtr, true) // ignore errors


    case "listscenes":
      st.PrintSceneList()

    case "listdevices":
      st.PrintDeviceList()

   

    case "switchon":
      st.DeviceSwitchOnOff(*namePtr, true)  

    case "switchoff":
      st.DeviceSwitchOnOff(*namePtr, false)  

    case "runscene":
      if(!st.ValidateScene(*namePtr)){
        fmt.Printf("[%s] is not a valid scene name.\nHere is the list\n\n",
                   *namePtr)
        st.PrintSceneList()
      }else{
        st.RunScene(*namePtr)
      }

    case "switchstatus":
      success, status := st.GetDeviceSwitchStatus(*namePtr)

      if(!success){
        fmt.Println("GetDeviceSwithcStatus failed")
      }else{
        fmt.Println(status.Switch.Value)
        fmt.Println(status.Switch.Timestamp)

        z, _ := status.Switch.Timestamp.Zone()
        fmt.Println("ZONE : ", z, " Time : ", status.Switch.Timestamp) // local time

        location, err := time.LoadLocation("EST")
        if err != nil {
            fmt.Println(err)
        }

        fmt.Println("ZONE : ", location, " Time : ", status.Switch.Timestamp.In(location)) // EST

      }

    default:
      help()
      os.Exit(2)

  }

  os.Exit(0)
     
}
