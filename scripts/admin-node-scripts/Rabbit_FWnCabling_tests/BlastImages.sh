#!/bin/bash

# This script is part of a package for testing the HPE NNF product, code named 'Rabbit' for customer LLNL.
# The code below flashes Rabbit-P & Rabbit-S parts that we know come from Benchmark with outdated images or no images at all
# The script is called with an argument that shall be called 'Chassis_num', which is the chassis number to be tested
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
#     * nC    = 1.8.3-13 (or higher)
#     * i210  = v3.255
#     * BIOS  = v0.3.0
#    Rabbit-S:
#     * uBoot = v1.11
#     * sFPGA = v2.05
#     * PAX   = v4.90, bld B464
#     * nC    = 1.8.3-13 (or higher)
#    E3.s Kioxia Drive FW = ver 1TCRS104
#========================================================================================================================
#   FUNCTIONS:
#     Wait4HSS - This function is called after issueing a reboot command to a nC. It's purpose is to wait until the nC
#                has fully rebooted and is responding to a PING.
#                The function takes a Chassis# (0-7) and slot# (4 or 7) as arguements
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
#========================================================================================================================
Chassis_num=$1

# Most of this flashing needs to have the main power rails OFF, so . . .
ssh x1002c$((Chassis_num))j7b0 node_auto_pd 0 force
ssh x1002c$((Chassis_num))j4b0 pax-power off

echo -e "\n\nStart by programming Rabbit-P i210 NIC FW and BIOS:\n"

echo -e "\nThe i210 image is shipped blank, so program it . . .\n"
scp $(pwd)/FwImages/i210_wnc_p2_sn01_n0.bin x1002c$((Chassis_num))j7b0:/rwfs/
ssh x1002c$((Chassis_num))j7b0 node_flashnet 0 /rwfs/i210_wnc_p2_sn01_n0.bin
# No easy way to verify successful flash now, wait to check in main Chassis_test.sh
echo -e "     NIC FW has been flashed, will verify later in the test"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
ssh x1002c$((Chassis_num))j7b0 rm /rwfs/i210_wnc_p2_sn01_n0.bin

echo -e "\nThe Rabbit-P node BIOS image is shipped blank, so program it . . .\n"
scp $(pwd)/FwImages/Exnnf.bios-0.3.0.ROM x1002c$((Chassis_num))j7b0:/rwfs/
ssh x1002c$((Chassis_num))j7b0 node_flashbios 0 /rwfs/Exnnf.bios-0.3.0.ROM
# No easy way to verify successful flash now, wait to check in main Chassis_test.sh
echo -e "     Node BIOS has been flashed, will verify later in the test"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
ssh x1002c$((Chassis_num))j7b0 rm /rwfs/Exnnf.bios-0.3.0.ROM

# Now, on to Rabbit-S which needs a couple of images blasted . . .

echo -e "\n\nNext we progam Rabbit-S with latest PCIe Switch FW & sFPGA:\n"
echo -e "\nStart with the PCIe Switch FW flashing . . .\n"
scp $(pwd)/FwImages/Prod_B464.data x1002c$((Chassis_num))j4b0:/rwfs/
# Note - one of the mtd's below is not flashable and exactly which one will vary. Trying to flash the non-flashable one won't hurt anything.
ssh x1002c$((Chassis_num))j4b0 flashcp -v /rwfs/Prod_B464.data /dev/mtd0
ssh x1002c$((Chassis_num))j4b0 flashcp -v /rwfs/Prod_B464.data /dev/mtd1
ssh x1002c$((Chassis_num))j4b0 flashcp -v /rwfs/Prod_B464.data /dev/mtd2
# No easy way to verify successful flash now, wait to check in main Chassis_test.sh
echo -e "     PCIe Switch FW has been flashed, will verify later in the test"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
ssh x1002c$((Chassis_num))j4b0 rm /rwfs/Prod_B464.data

echo -e "Now start programming Rabbit-S sFPGA (if necessary) . . .\n"

if !(ssh x1002c$((Chassis_num))j4b0 dmesg | grep --quiet 'sFPGA-RBTS rev 205'); then
    echo -e "\nRabbit-S sFPGA is not v2.05, starting the flash"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    scp $(pwd)/FwImages/sfpga_rbts_21_09_ver02_05.bin x1002c$((Chassis_num))j4b0:/rwfs/
    ssh x1002c$((Chassis_num))j4b0 fpga_flasher -f /rwfs/sfpga_rbts_21_09_ver02_05.bin
    ssh x1002c$((Chassis_num))j4b0 rm /rwfs/sfpga_rbts_21_09_ver02_05.bin
    ssh x1002c$((Chassis_num))j4b0 do_fpga_reboot -f
    echo -e "\n\nFlashing done, now need to reboot the sFPGA. Going to need to wait a couple of minutes for sFPGA to boot back up . . .\n"
    row_num=4
    Wait4HSS
    if [ $? -ne 0 ]; then
       exit -1
    fi

    echo -e "\n\nNow lets see if it's back up - we'll PING the nC:\n"
    if !(ping x1002c$((Chassis_num))j4b0 -c 1); then
        echo -e "\n\nRabbit-S nC should have booted by now - we've got a problem!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
        echo "Exiting test!"  | tee -a Chassis$((Chassis_num)).log
        exit -1
    fi
    echo -e "\n\nReboot done, now need to verify proper version . . .\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
    if !(ssh x1002c$((Chassis_num))j4b0 dmesg | grep --quiet 'sFPGA-RBTS rev 205'); then
        echo -e "sFPGA failed version check again - programming didn't work.\n\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
        echo "Exiting test!"  | tee -a Chassis$((Chassis_num)).log
        exit -1
    fi
    echo -e "     Rabbit-S sFPGA has been flashed and verified!\n"  | tee -a $(pwd)/logs/Chassis$((Chassis_num)).log
fi
