#!/bin/bash

# This script is for testing the HPE NNF product, code named 'Rabbit'.
# It tests low-level FW versions, presence and seating of cables, and ability of HW to power-on and successfully boot.
# The script is called with an argument which shall be called 'Chassis_num', which is the chassis number to be tested
# The following assumtions are made/expected coming into this script:
#   * The Rabbits are installed in the Cabinet with geo-location "x1002c?j4b0", "x1002c?j7b0", and "x1002c?j7b0n0"
#   * The Rabbits are in the HPCM database & have had their SSH keys set
#   * The Rabbits nC's are fully booted
#
# ============ This is Ver1.1 of the script, 6/30/2023 ===========
#  This script expects and/or programs the following FW versions:
#    Rabbit-P:
#     * uBoot = v1.11
#     * nFPGA = v2.09
#     * nC    = v1.8.3-13 (or higher)
#     * i210  = v3.25
#     * BIOS  = v0.3.0
#    Rabbit-S:
#     * uBoot = v1.11
#     * sFPGA = v2.05
#     * nC    = v1.8.3-13 (or higher)
#     * PAX   = v4.90, bld B464
#    E3.s Kioxia Drive FW = ver 1TCRS104
#    Slingshot FW = 1.5.41
#===================================================================
Chassis_num=$1                # This is the calling-script-supplied arguement of '0' thru '7'
DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
echo -e "Chassis-$((Chassis_num)) begin testing at $DATE_TIME" > $(pwd)/logs/Chassis$((Chassis_num)).log

#========================================================================================================================
#   FUNCTIONS:
#     Wait4HSS - This function is called after issuing a reboot command to a nC. It's purpose is to wait until the nC
#                has fully rebooted and is responding to a PING.
#                The function takes a Chassis# (0-7) and slot# (4 or 7) as arguements
#     PowerUpNode - This function is called after issuing a Node Power-up command. It's purpose is to wait until the
#                Node has fully PXE booted and is responding to a PING.
#                The function takes a Chassis# (0-7) as an argument 
#     GetPAXLinks - This function gets the link-status of each of the 2 PAXs and makes sure each PAX is linked with 8
#                E3.s drives. It them power-cycles the Rabbit-S and re-performs the link checks.
#========================================================================================================================
function Wait4HSS {
    sleep 2m
    delay=4
        while [ $delay -gt 0 ];
        do
          if [ $delay == 1 ]; then
             echo -e "\n\nRabbit nC in Chassis$Chassis_num , Row $row_num should have booted by now - we've got a problem!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
             echo -e "\n\nExiting test!"
             exit -1
          elif !(ping x1002c$((Chassis_num))j$((row_num))b0 -c 1); then
             echo -e "\nNot responding yet, we'll give it another minute . . .\n"
             sleep 1m
          else
             delay=1
          fi
          ((delay--))
        done

        echo -e "\n\nNow lets see if it's back up - we'll PING the nC:\n\n"
        if !(ping x1002c$((Chassis_num))j$((row_num))b0 -c 1); then
           echo -e "\n\nRabbit nC in Chassis$Chassis_num , Row $row_num should have booted by now - we've got a problem!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
           echo -e "\n\nExiting test!"
           exit -1
        fi
}

function PowerUpNode {
    echo -e "\nCheck Chassis$Chassis_num Rabbit node power status . . ."  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log

    # See if the node is already powered-up, of so skip the long delays . . .
    if ! (cm power status -n x1002c$((Chassis_num))j7b0n0 | grep --quiet 'BOOTED'); then
        echo -e "     Chassis$Chassis_num Rabbit node powering on . . ."  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
        ssh x1002c$((Chassis_num))j7b0 node_auto_pu 0
        echo -e "\n\nRabbit node PXE-booting, this will take 5-10 minutes . . .\n\n"
        sleep 5m

    # After the initial 5min delay, give it up to 6more min to PXE boot . . .
        boot_dly=7
            while [ $boot_dly -gt 0 ];
            do
               if [ $boot_dly == 1 ]; then
                  echo -e "\n\nRabbit-P node should have booted by now - we've got a problem!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
                  echo "Exiting test"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
                  exit -1
               elif ! (ping x1002c$((Chassis_num))j7b0n0 -c 1); then
                  echo -e "\n\nNot responding yet, we'll give it another minute . . .\n\n"
                  sleep 1m
               else
                  boot_dly=1
               fi
               ((boot_dly--))
             done
        echo -e "\n\nRabbit's NIC is PING'ing, so almost there!"
        sleep 0.6m      # OS still needs another ~30sec before ready for commands

    else
        echo -e "\nRabbit node is already booted up!"

    fi
}

function GetPAXLinks {
    rm $(pwd)/logs/Chassis$((Chassis_num))_pax*                                             # Get rid of old logs so we start fresh
    ssh x1002c$((Chassis_num))j4b0 rm /root/pax*                                            # Clear out any old logs
    scp $(pwd)/PAX_logs.sh x1002c$((Chassis_num))j4b0:/root/
    ssh x1002c$((Chassis_num))j4b0 ./PAX_logs.sh                                            # Clear out any old logs
    sleep 3s                                                                                # Give it a few seconds
    scp x1002c$((Chassis_num))j4b0:/root/pax0.log $(pwd)/logs/Chassis$((Chassis_num))_pax0.log   # Grab the PAX0 log
    scp x1002c$((Chassis_num))j4b0:/root/pax1.log $(pwd)/logs/Chassis$((Chassis_num))_pax1.log   # Grab the PAX1 log
  
 # --- Next we delete unwanted lines of text and link-status' from the file
    sed -i '/GFMS/d' $(pwd)/logs/Chassis$((Chassis_num))_pax0.log                           # Get rid of all the GFMS events
    sed -i '/x16/d' $(pwd)/logs/Chassis$((Chassis_num))_pax0.log                            # Get rid of the x16 ports
    sed -i '/stk:4/d' $(pwd)/logs/Chassis$((Chassis_num))_pax0.log                          # Get rid of any upstream links (servers)
    sed -i '/stk:5/d' $(pwd)/logs/Chassis$((Chassis_num))_pax0.log                          # Get rid of any upstream links (servers)

    sed -i '/GFMS/d' $(pwd)/logs/Chassis$((Chassis_num))_pax1.log                           # Get rid of all the GFMS events
    sed -i '/x16/d' $(pwd)/logs/Chassis$((Chassis_num))_pax1.log                            # Get rid of the x16 ports
    sed -i '/stk:4/d' $(pwd)/logs/Chassis$((Chassis_num))_pax1.log                          # Get rid of any upstream links (servers)
    sed -i '/stk:5/d' $(pwd)/logs/Chassis$((Chassis_num))_pax1.log                          # Get rid of any upstream links (servers)

  # --- Now we exit w/2 files; the PAX0 file has its 9 link-status' and the PAX1 file has its 9 link status'
}

#========================================================================================================================
#========================================================================================================================

# Make sure Rabbits are present and HSS is booted/on-network . . .
     echo -e "\n\n Lets make sure the nC's are booted and on the network . . . \n\n"

     if !(ping x1002c$((Chassis_num))j7b0 -c 1); then           # Can we ping the nC?
        DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
        echo -e "\nRabbit-P nC in Chassis$Chassis_num isn't reachable, somethings wrong! \nExiting test at $DATE_TIME!"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
        cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
        exit -1
     fi
     if !(ping x1002c$((Chassis_num))j4b0 -c 1); then           # Can we ping the nC?
        DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
        echo -e "\nRabbit-S nC in Chassis1 isn't reachable, somethings wrong! \nExiting test at $DATE_TIME!"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
        cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
        exit -1
     fi
     echo -e "\n\nnC's are responding!"
#========================================================================================================
# All Rabbits present & responding; before we bring the node up we'll flash some blank parts & some
#      FW images we know have changed . . .
   echo -e "\nStart flashing known blank & down-rev parts . . ."  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
   ./BlastImages.sh $Chassis_num
   if [ $? -ne 0 ]; then
     DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
     echo -e "\nBye-bye from BlastImages at $DATE_TIME!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
     cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
     exit -1
   fi
   echo -e "Blank & down-rev images programmed!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log


#========================================================================================================
# Now we'll verify that the Rabbit-S HSS (uBoot, nC, and sFPGA) has the most recent FW images
#
had2flash="no"

# Update the uBoot FW if not currently 1.11
if !(ssh x1002c$((Chassis_num))j4b0 dmesg | grep --quiet 'cray_uboot_ver=1.11'); then
    echo "Problem: Rabbit-S uBoot is down-rev, need to Flash it!"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    scp $(pwd)/FwImages/u-boot_wnc_1.11.bin x1002c$((Chassis_num))j4b0:/rwfs/
    ssh x1002c$((Chassis_num))j4b0 uboot_flasher /rwfs/u-boot_wnc_1.11.bin
    ssh x1002c$((Chassis_num))j4b0 rm /rwfs/u-boot_wnc_1.11.bin
    had2flash="yes"
fi

# Update the sFPGA if not currently 2.05
if !(ssh x1002c$((Chassis_num))j4b0 dmesg | grep --quiet 'sFPGA-RBTS rev 205'); then
    echo -e "\nProblem - Rabbit-S sFPGA is not v2.05! Lets try flashing it:"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    scp $(pwd)/FwImages/sfpga_rbts_21_09_ver02_05.bin x1002c$((Chassis_num))j4b0:/rwfs/
    ssh x1002c$((Chassis_num))j4b0 fpga_flasher -f /rwfs/sfpga_rbts_21_09_ver02_05.bin
    ssh x1002c$((Chassis_num))j4b0 rm /rwfs/sfpga_rbts_21_09_ver02_05.bin
    had2flash="yes"
fi

# Update the nC FW if not currently 1.8.3-13. This FW revision has necessary fixes for Rabbit.
if !(ssh x1002c$((Chassis_num))j4b0 cat /usr/etc/version/hms-controllers.version | grep '1.8.3-13'); then
    echo "Problem: Rabbit-S nC is down-rev, need to Flash it!"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    scp $(pwd)/FwImages/nc-1.8.3-13.itb x1002c$((Chassis_num))j4b0:/root/
    scp $(pwd)/Prgrm_nc.sh x1002c$((Chassis_num))j4b0:/root/
    ssh x1002c$((Chassis_num))j4b0 ./Prgrm_nc.sh
    had2flash="yes"
    row_num=4
    Wait4HSS
    if [ $? -ne 0 ]; then
       DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
       echo -e "\nBye-bye from Rabbit-P nC Wait4HSS at $DATE_TIME!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
       cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
       ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
       ssh x1002c$((Chassis_num))j4b0 pax-power off
       exit -1
    fi
fi

# If any uBoot or sFPGA images in Rabbit-S' were flashed, then we need to reboot the HSS subsystem
if [ "$had2flash" == "yes" ] ; then
    echo -e "\nRebooting Rabbit-S HSS subsystem, will have to wait a few minutes"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    ssh x1002c$((Chassis_num))j4b0 reboot
    row_num=4
    Wait4HSS
    if [ $? -ne 0 ]; then
       DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
       echo -e "\nBye-bye from Rabbit-S HSS Wait4HSS at $DATE_TIME!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
       cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
       ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
       ssh x1002c$((Chassis_num))j4b0 pax-power off
       exit -1
    fi

    echo -e "\nReboot done, now need to verify proper versions . . .\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    if !(ssh x1002c$((Chassis_num))j4b0 dmesg | grep --quiet 'cray_uboot_ver=1.11'); then
      DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
      echo -e "\nRabbit-S uBoot failed version-check again at $DATE_TIME!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
      cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
      ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
      ssh x1002c$((Chassis_num))j4b0 pax-power off
      exit -1
    fi
    echo -e "     Rabbit-S uBoot is v1.11!"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log

    if !(ssh x1002c$((Chassis_num))j4b0 cat /usr/etc/version/hms-controllers.version | grep '1.8.3-13'); then
      DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
      echo -e "\nRabbit-S nC FW failed version-check again at $DATE_TIME!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
      cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
      ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
      ssh x1002c$((Chassis_num))j4b0 pax-power off
      exit -1
    fi
    echo -e "     Rabbit-S nC FW is v1.8.3-13!"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log

    if !(ssh x1002c$((Chassis_num))j4b0 dmesg | grep --quiet 'sFPGA-RBTS rev 205'); then
      DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
      echo -e "\nRabbit-S sFPGA failed version-check again at $DATE_TIME!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
      cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
      ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
      ssh x1002c$((Chassis_num))j4b0 pax-power off
      exit -1
    fi
    echo -e "     Rabbit-S sFPGA is v2.05!"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
fi
echo -e "Rabbit-S HSS subsystem FW is good-to-go!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log

#========================================================================================================
# Now we'll verify that the Rabbit-P HSS (uBoot, nC, and nFPGA) has the most recent FW images
#
had2flash="no"

# Update the uBoot FW if not currently 1.11
if !(ssh x1002c$((Chassis_num))j7b0 dmesg | grep --quiet 'cray_uboot_ver=1.11'); then
    echo "Problem: Rabbit-P uBoot is down-rev, need to Flash it!"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    scp $(pwd)/FwImages/u-boot_wnc_1.11.bin x1002c$((Chassis_num))j7b0:/rwfs/
    ssh x1002c$((Chassis_num))j7b0 uboot_flasher /rwfs/u-boot_wnc_1.11.bin
    ssh x1002c$((Chassis_num))j7b0 rm /rwfs/u-boot_wnc_1.11.bin
    had2flash="yes"
fi

# Update the nFPGA if not currently 2.09
if !(ssh x1002c$((Chassis_num))j7b0 dmesg | grep --quiet 'nFPGA-RBTP rev 209'); then
    echo -e "\nProblem - Rabbit-P nFPGA is not v2.09! Lets try flashing it:"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    scp $(pwd)/FwImages/nfpga_rbtp_24_09_ver02_09.bin x1002c$((Chassis_num))j7b0:/rwfs/
    ssh x1002c$((Chassis_num))j7b0 fpga_flasher -f /rwfs/nfpga_rbtp_24_09_ver02_09.bin 
    ssh x1002c$((Chassis_num))j7b0 rm /rwfs/nfpga_rbtp_24_09_ver02_09.bin 
    had2flash="yes"
fi

# Update the nC FW if not currently 1.8.3-13. This FW revision has necessary fixes for Rabbit.
if !(ssh x1002c$((Chassis_num))j7b0 cat /usr/etc/version/hms-controllers.version | grep '1.8.3-13'); then
    echo "Problem: Rabbit-P nC is down-rev, need to Flash it!"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    scp $(pwd)/FwImages/nc-1.8.3-13.itb x1002c$((Chassis_num))j7b0:/root/
    scp $(pwd)/Prgrm_nc.sh x1002c$((Chassis_num))j7b0:/root/
    ssh x1002c$((Chassis_num))j7b0 ./Prgrm_nc.sh
    had2flash="yes"
    row_num=7
    Wait4HSS
    if [ $? -ne 0 ]; then
       DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
       echo -e "\nBye-bye from Rabbit-P nC Wait4HSS at $DATE_TIME!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
       cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
       ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
       ssh x1002c$((Chassis_num))j4b0 pax-power off
       exit -1
    fi
fi

# If any images in Rabbit-P' were flashed, then we need to reboot the HSS subsystem
if [ "$had2flash" == "yes" ] ; then
    echo -e "Rebooting Rabbit-P HSS subsystem, will have to wait a few minutes"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    ssh x1002c$((Chassis_num))j7b0 reboot
    row_num=7
    Wait4HSS
    if [ $? -ne 0 ]; then
       DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
       echo -e "\nBye-bye from Rabbit-P HSS Wait4HSS at $DATE_TIME!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
       cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
       ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
       ssh x1002c$((Chassis_num))j4b0 pax-power off
       exit -1
    fi

    echo -e "Reboot done, now need to verify proper versions . . ."  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    if !(ssh x1002c$((Chassis_num))j7b0 dmesg | grep --quiet 'cray_uboot_ver=1.11'); then
      DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
      echo -e "\nRabbit-P uBoot failed version-check again at $DATE_TIME!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
      cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
      ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
      ssh x1002c$((Chassis_num))j4b0 pax-power off
      exit -1
    fi
    echo -e "     Rabbit-P uBoot is v1.11!"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log

    if !(ssh x1002c$((Chassis_num))j7b0 cat /usr/etc/version/hms-controllers.version | grep '1.8.3-13'); then
      DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
      echo -e "\nRabbit-P nC FW failed version-check again at $DATE_TIME!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
      cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
      ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
      ssh x1002c$((Chassis_num))j4b0 pax-power off
      exit -1
    fi
    echo -e "     Rabbit-P nC FW is v1.8.3-13!"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log

    if !(ssh x1002c$((Chassis_num))j7b0 dmesg | grep --quiet 'nFPGA-RBTP rev 209'); then
      DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
      echo -e "\nRabbit-P nFPGA failed version-check again at $DATE_TIME!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
      cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
      ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
      ssh x1002c$((Chassis_num))j4b0 pax-power off
      exit -1
    fi
    echo -e "     Rabbit-P nFPGA is v2.09!"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
fi
echo -e "Rabbit-P HSS subsystem FW is good-to-go!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log

#========================================================================================================
# Power-up the PAX's on Rabbit-S and make sure they link with all 16 drives. Then power-cycle once and 
#    re-verify that all 16 drives were found. . .

   echo -e "Turn-on the Rabbit-S to check link-status to drives."  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
   ssh x1002c$((Chassis_num))j4b0 pax-power on
   sleep 10s
   GetPAXLinks                                                                               # This function outputs 2 files; PAX0 & PAX1, each with
                                                                                             #    each with the respective PAX's link status'
   declare -i num_drv_pax0=$(cat $(pwd)/logs/Chassis$((Chassis_num))_pax0.log | grep -c -F 'neg[x04]')
   echo -e "   Number of drives found on PAX0 is $num_drv_pax0 "  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
   if (($num_drv_pax0 != 8 )); then                                                          # Should find 8 Drives
      DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                                                 # Date/time stamp for the log file
      echo -e "\nNot linking with all 8 drives on PAX0. Failed test at $DATE_TIME!\n" | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
      cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
      cat $(pwd)/logs/Chassis$((Chassis_num))_pax0.log >> $(pwd)/logs/Chassis$((Chassis_num))_Failure.log  # Append for analysis
      ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
      ssh x1002c$((Chassis_num))j4b0 pax-power off
      exit -1
   fi
   echo -e "\nPAX0 is linking with all 8 drives on 1st power-cycle!\n"

   declare -i num_drv_pax1=$(cat $(pwd)/logs/Chassis$((Chassis_num))_pax1.log | grep -c -F 'neg[x04]')
   echo -e "   Number of drives found on PAX1 is $num_drv_pax1 "  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
   if (($num_drv_pax1 != 8 )); then                                                          # Should find 8 Drives
      DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                                                 # Date/time stamp for the log file
      echo -e "\nNot linking with all 8 drives on PAX1. Failed test at $DATE_TIME!\n" | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
      cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
      cat $(pwd)/logs/Chassis$((Chassis_num))_pax1.log >> $(pwd)/logs/Chassis$((Chassis_num))_Failure.log  # Append for analysis
      ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
      ssh x1002c$((Chassis_num))j4b0 pax-power off
      exit -1
   fi
   echo -e "\nPAX1 is linking with all 8 drives on 1st power-cycle!\n"

   ssh x1002c$((Chassis_num))j4b0 pax-power off                                             # Power-cycle PAX's and repeat
   sleep 10s
#------------------------------------------------
   echo -e "Power-cycling Rabbit-S to check link-status a 2nd time."  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
   ssh x1002c$((Chassis_num))j4b0 pax-power on
   sleep 10s
   GetPAXLinks                                                                               # This function outputs 2 files; PAX0 & PAX1, each with
                                                                                             #    each with the respective PAX's link status'
   declare -i num_drv_pax0=$(cat $(pwd)/logs/Chassis$((Chassis_num))_pax0.log | grep -c -F 'neg[x04]')
   echo -e "   Number of drives found on PAX0 is $num_drv_pax0 "  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
   if (($num_drv_pax0 != 8 )); then                                                          # Should find 8 Drives
      DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                                                 # Date/time stamp for the log file
      echo -e "\nNot linking with all 8 drives on PAX0. Failed test at $DATE_TIME!\n" | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
      cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
      cat $(pwd)/logs/Chassis$((Chassis_num))_pax0.log >> $(pwd)/logs/Chassis$((Chassis_num))_Failure.log  # Append for analysis
      ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
      ssh x1002c$((Chassis_num))j4b0 pax-power off
      exit -1
   fi
   echo -e "\nPAX0 is linking with all 8 drives on 2nd power-cycle too!\n"

   declare -i num_drv_pax1=$(cat $(pwd)/logs/Chassis$((Chassis_num))_pax1.log | grep -c -F 'neg[x04]')
   echo -e "   Number of drives found on PAX1 is $num_drv_pax1 "  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
   if (($num_drv_pax1 != 8 )); then                                                          # Should find 8 Drives
      DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                                                 # Date/time stamp for the log file
      echo -e "\nNot linking with all 8 drives on PAX1. Failed test at $DATE_TIME!\n" | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
      cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
      cat $(pwd)/logs/Chassis$((Chassis_num))_pax1.log >> $(pwd)/logs/Chassis$((Chassis_num))_Failure.log  # Append for analysis
      ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
      ssh x1002c$((Chassis_num))j4b0 pax-power off
      exit -1
   fi
   echo -e "\nPAX1 is linking with all 8 drives on 2nd power-cycle too!\n"

#========================================================================================================
# Time to power-on the nodes on & make sure they successfully PXE boot . . .

# Start by turning all nodes on . . .
   PowerUpNode                    # Then wait/verify the PXE boot
   if [ $? -ne 0 ]; then
     DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
     echo -e "\nBye-bye from PowerUpNode at $DATE_TIME!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
     cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
     ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
     ssh x1002c$((Chassis_num))j4b0 pax-power off
     exit -1
   fi
   echo -e "Rabbit is powered-up, now start testing . . .\n"   | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
#========================================================================================================
# With node powered up & PXE booted we can begin testing . . .

echo -e "Now on to FW checks:"   | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log

sleep 1m

#--- Check the Nodes BIOS now . . .
if !(ssh x1002c$((Chassis_num))j7b0n0 dmidecode | grep --quiet '0.3.0'); then
    echo "Problem: BIOS is not v0.3.0 - need to Flash it!"
    ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
    sleep 0.3s
    scp $(pwd)/FwImages/Exnnf.bios-0.3.0.ROM x1002c$((Chassis_num))j7b0:/rwfs/
    ssh x1002c$((Chassis_num))j7b0 node_flashbios 0 Exnnf.bios-0.3.0.ROM
    ssh x1002c$((Chassis_num))j7b0 rm Exnnf.bios-0.3.0.ROM
    PowerUpNode
    if [ $? -ne 0 ]; then
     DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
      echo -e "\nBye-bye from PowerUpNode at Rabbit-P BIOS at $DATE_TIME!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
      cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
      ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
      ssh x1002c$((Chassis_num))j4b0 pax-power off
      exit -1
    fi

    echo -e "\nReboot done, now need to verify proper BIOS version . . .\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    if !(ssh x1002c$((Chassis_num))j7b0n0 dmidecode | grep --quiet 'Version: 0.3.0'); then
      echo -e "Rabbit in Chassis $Chassis_num failed BIOS version check again - need to debug so exiting the test\n\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
      cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
      ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
      ssh x1002c$((Chassis_num))j4b0 pax-power off
      exit -1
    fi
fi
echo "     Rabbit-P's BIOS is v0.3.0!"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
sleep 0.3s      # Need delay here - sometimes back-to-back SSH's get rejected

#--- Check the Rabbit-P embedded i210 NIC . . .
if !(ssh x1002c$((Chassis_num))j7b0n0 ethtool -i enp193s0 | grep --quiet '3.25'); then
    echo -e "Problem: i210 NIC is missing or has wrong FW version - need to flash it!"

    ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
    sleep 0.1m                                         # Delay so power-off can complete
    scp $(pwd)/FwImages/i210_wnc_p2_sn01_n0.bin x1002c$((Chassis_num))j7b0:/rwfs/
    ssh x1002c$((Chassis_num))j7b0 node_flashnet 0 i210_wnc_p2_sn01_n0.bin
    ssh x1002c$((Chassis_num))j7b0 rm i210_wnc_p2_sn01_n0.bin
    ssh x1002c$((Chassis_num))j7b0 node_auto_pu 0
    echo -e "NIC FW has been flashed, powering Rabbit-P back on now so we'll wait a few minutes . . ."  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    PowerUpNode
    if [ $? -ne 0 ]; then
       DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
       echo -e "\nBye-bye from the i210 NICs PowerUpNode at $DATE_TIME!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
       cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
       ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
       ssh x1002c$((Chassis_num))j4b0 pax-power off
       exit -1
    fi

    echo -e "\nReboot done, now need to verify proper version . . .\n"
    if !(ssh x1002c$((Chassis_num))j7b0n0 ethtool -i enp193s0 | grep --quiet '3.25'); then
      DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
      echo -e "\nThe i210 NIC failed version check again at $DATE_TIME!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
      cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
      ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
      ssh x1002c$((Chassis_num))j4b0 pax-power off
      exit -1
    fi
fi
echo -e "     i210 NIC has FW v3.25!"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log

#--- Now check the Rabbit-S PAX chips . . .
Need2FlshPAXs=0
if !(ssh x1002c$((Chassis_num))j7b0n0 switchtec fw-info /dev/switchtec0 | grep --quiet 'Version: 4.90 B464'); then
	echo "PAX0 has wrong FW version, need to flash!"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
        Need2FlshPAXs=1
fi

if !(ssh x1002c$((Chassis_num))j7b0n0 switchtec fw-info /dev/switchtec1 | grep --quiet 'Version: 4.90 B464'); then
	echo "PAX1 has wrong FW version, need to flash!"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
        Need2FlshPAXs=1
fi

if [ $Need2FlshPAXs = '1' ]; then
    echo "Turning Rabbit off to start flashing process . . ."
    cm power shutdown -n x1002c$((Chassis_num))j7b0n0              # Turn node off
    ssh x1002c$((Chassis_num))j4b0 pax-power off              # Turn off main rails of Rabbit-S
    sleep 0.3s
    scp $(pwd)/FwImages/Prod_B464.data x1002c$((Chassis_num))j4b0:/rwfs/
# Note - one of the mtd's below is not flashable and exactly which one will vary. Trying to flash the non-flashable one won't hurt anything.
    ssh x1002c$((Chassis_num))j4b0 flashcp -v /rwfs/Prod_B464.data /dev/mtd0
    ssh x1002c$((Chassis_num))j4b0 flashcp -v /rwfs/Prod_B464.data /dev/mtd1
    ssh x1002c$((Chassis_num))j4b0 flashcp -v /rwfs/Prod_B464.data /dev/mtd2
    ssh x1002c$((Chassis_num))j4b0 rm /rwfs/Prod_B464.data
    ssh x1002c$((Chassis_num))j7b0 node_auto_pu 0
    echo -e "Rabbit-S PAXs have been flashed, power Rabbit back on now so we'll need to wait a few minutes . . ."  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    PowerUpNode
    if [ $? -ne 0 ]; then
       DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
       echo -e "\nBye-bye from the PAXs PowerUpNode at $DATE_TIME!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
       cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
       ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
       ssh x1002c$((Chassis_num))j4b0 pax-power off
       exit -1
    fi

    echo -e "PAX flashing done & rebooted, now we'll check version again . . ."  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log

    Need2FlshPAXs=0
    if !(ssh x1002c$((Chassis_num))j7b0n0 switchtec fw-info /dev/switchtec0 | grep --quiet 'Version: 4.90 B464'); then
       Need2FlshPAXs=1
    fi
    if !(ssh x1002c$((Chassis_num))j7b0n0 switchtec fw-info /dev/switchtec1 | grep --quiet 'Version: 4.90 B464'); then
       Need2FlshPAXs=1
    fi
fi
#--- Final check
if [ $Need2FlshPAXs = '1' ]; then
      DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
      echo -e "\nThe PAXs failed version check again at $DATE_TIME!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
      cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
      ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
      ssh x1002c$((Chassis_num))j4b0 pax-power off
      exit -1
fi
	
echo "     Rabbit-S PAX0 and PAX1 have FW v4.90 B464!"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log

#Finally, verify &/or flash the E3.s drives . . .
echo -e "\nNow lets verify drives have the correct FW version . . ."  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
./updateDriveFirmware.sh x1002c$((Chassis_num))j7b0n0 1TCRS104 /root/1TCRS104.std
   if [ $? -ne 0 ]; then
     cat $(pwd)/logs/x1002c$((Chassis_num))j7b0n0.log >> $(pwd)/logs/Chassis$((Chassis_num)).log
     rm $(pwd)/logs/x1002c$((Chassis_num))j7b0n0.log
     DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
     echo -e "\nBye-bye from UpdateDriveFW at $DATE_TIME!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
     cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
     exit -1
   fi
cat $(pwd)/logs/x1002c$((Chassis_num))j7b0n0.log >> $(pwd)/logs/Chassis$((Chassis_num)).log
rm $(pwd)/logs/x1002c$((Chassis_num))j7b0n0.log

#===============================================================================================
#--- Check to make sure proper CPU & DIMMs installed . . .
if !(ssh x1002c$((Chassis_num))j7b0n0 dmidecode | grep 'AMD EPYC 7713'); then
    DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
    echo -e "\nCPU is not an AMD EPYC 7713 - replace with correct CPU. Failed test at $DATE_TIME!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
    ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
    ssh x1002c$((Chassis_num))j4b0 pax-power off
    exit -1
fi
if !(ssh x1002c$((Chassis_num))j7b0n0 dmidecode -t 17 |egrep 'Size: 32' |egrep -v V |wc -l |egrep 8); then
    DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
    echo -e "\nNot recognizing 256GB of Memory - bad or missing DIMM. Failed test at $DATE_TIME!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
    ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
    ssh x1002c$((Chassis_num))j4b0 pax-power off
    exit -1
fi
echo -e "CPU & DDR4 Memory check good!"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log

#--- Check to make sure both NICs present & linking at 200Gb. . .
declare -i num_nics=$(ssh x1002c$((Chassis_num))j7b0n0 cxi_stat | grep -c 'cxi')
#echo -e "\n\nNumber of nics found is $num_nics \n\n"
if (($num_nics != 2 )); then                                                   # Should fund 2 Devices
    DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
    echo -e "\nNot seeing both HSN NICs; bad Sawtooth or Rabbit-P. Failed test at $DATE_TIME!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
    ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
    ssh x1002c$((Chassis_num))j4b0 pax-power off
    exit -1
fi
echo -e "Two HSN NICs found!"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log

declare -i num_200Gb=$(ssh x1002c$((Chassis_num))j7b0n0 cxi_stat | grep -c 'Link speed: BS_200G')
#echo -e "\n\nNumber of nic links found at 200Gb is $num_nics \n\n"
if (($num_200Gb != 2 )); then                                                   # Should fund 2 links @ 200Gbit
    DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
    echo -e "\nHSN NICs not linking at 200Gb; check Sawtooth & cabling. Failed test at $DATE_TIME!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
    ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
    ssh x1002c$((Chassis_num))j4b0 pax-power off
    exit -1
fi

echo -e "Both HSN NICs present & linking at 200Gb!"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log

#===============================================================================================
echo -e "\nNow lets check to make sure the Rabbits internal cables are plugged in & seated:"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log

if !(ssh x1002c$((Chassis_num))j7b0 ip a | grep --quiet 'ppp0'); then  #The ppp0 interface is the link thru the Ribbon Cable
    DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
    echo -e "\n\nRabbit Interconnect not linking at $DATE_TIME - check Ribbon cable (P3 on Rabbit-P)"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
    ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
    ssh x1002c$((Chassis_num))j4b0 pax-power off
    exit -1
fi
echo -e "     Rabbit Interconnect Ribbon Cable is plugged in!"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log

if !(ssh x1002c$((Chassis_num))j7b0n0 lspci | grep --quiet '4200'); then  # Device ID 4200 are the PAX's - if neither show up then probablly the HSS link
    DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
    echo -e "\n\nNeither PAX linking at $DATE_TIME - check x8 Rabbit Mgmt cable (J3000 on Rabbit-P)"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
    ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
    ssh x1002c$((Chassis_num))j4b0 pax-power off
    exit -1
fi
echo "     Rabbit x8 Management Cable is plugged in!"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log


if (ssh x1002c$((Chassis_num))j7b0n0 lspci -s 03:00.1 -vvv | grep --quiet 'downgraded'); then  #Bus3, Dev0, Fn1 is the PAX0
    DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
    echo -e "\n\nPAX0 link is x8 at $DATE_TIME - check the following:"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    echo -e "    1) Cables to Rabbit-P J3002 and J3003 aren't backwards."  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    echo -e "    2) Cable not seated properly between Rabbit-P, J3002, and Rabbit-S, J7101."  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    echo -e "    3) Cable not seated properly between Rabbit-P, J3003, and Rabbit-S, J7100."  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
    ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
    ssh x1002c$((Chassis_num))j4b0 pax-power off
    exit -1
fi
echo "     CPU-to-PAX0 cables are correct!"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log

if (ssh x1002c$((Chassis_num))j7b0n0 lspci -s 81:00.1 -vvv | grep --quiet 'downgraded'); then  #Bus81, Dev0, Fn1 is the PAX1
    DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
    echo -e "\n\nPAX1 link is x8 at $DATE_TIME - check the following:"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    echo -e "    1) Cables to Rabbit-P J3004 and J3005 aren't backwards."  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    echo -e "    2) Cable not seated properly between Rabbit-P, J3004, and Rabbit-S, J7102."  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    echo -e "    3) Cable not seated properly between Rabbit-P, J3005, and Rabbit-S, J7103."  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    cp $(pwd)/logs/Chassis$((Chassis_num)).log $(pwd)/logs/Chassis$((Chassis_num))_Failure.log
    ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
    ssh x1002c$((Chassis_num))j4b0 pax-power off
    exit -1
fi
echo -e "     CPU-to-PAX1 cables are correct!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
#=================================================================================================
# Testing complete
DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
echo "FW & Cabling Tests PASSED on Chassis-$Chassis_num at $DATE_TIME!!!!"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log

