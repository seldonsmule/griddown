# griddown
This program talks to a Telsa Powerwall and determines if the power grid is working or the system is acting as a backup.  When power loss is detected, it will execute a predefined SmartThings scene.  When power is restored, it will execute another predefined SmartThings scene.

# Dependencies
- Smartthings account and compatible hub
- Telsa Powerwall and login information.  Note, this can only be done within the network of the installed system - not accessbile on the internet
- This command is intended to run from cron as it is currently written

# Powerwall Requirements
- Tesla Powerwall and gateway
- Know the IP address of the gateway (intenal webserver)
- Create a local login with password on the local gateway's web interface
- The provided powerwall.cer assumes you have a DNS (or host file) with the name powerwall pointing to the IP address assigned to your powerwall system in your house

# SmartThings Requirements
- Samsung SmartThings account
- SmartThings compatible hub
- Personal Access Token -> https://account.smartthings.com/tokens
- A pre-created scene to take actions when power is lost
-- Example: GridDown -> Turn off water heater
- A pre-created scene to take action when power is restored
-- Example: GridUp -> Turn on water heater

