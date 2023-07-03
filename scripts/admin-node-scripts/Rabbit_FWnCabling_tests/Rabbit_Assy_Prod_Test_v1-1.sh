#!/bin/bash

# This script is for testing the HPE NNF product, code named 'Rabbit'.
# It tests low-level FW versions, presence and seating of cables, and ability of HW subsytems to power-on and successfully boot.
# It is invoked with an argument which is one of three types:
#   * A single number, 0-7, representing the single Chassis # to be tested
#   * An argument of 'tds' which will cause the testing of Chassis #'s 1 and 3
#   * An argument of 'all' which will cause the testing of all 8 Chassis, Chassis #'s 0 thru 7
# Valid examples of calling this script are:
#   * ./Rabbit_Assy_Prod_Test.sh 6     ==> Tests the Rabbit in Chassis #6 
#   * ./Rabbit_Assy_Prod_Test.sh tds   ==> Tests the Rabbits in Chassis #1 and #3 
#   * ./Rabbit_Assy_Prod_Test.sh all   ==> Tests the Rabbits in all Chassis, #0 thru #7
# This scripts expects to have FW images (binary, .data, .ROM) in the /scripts/ directory
# This script puts testing output into a log file located in the /logs/ directory
# Should a failure be found a 'Failure' log will be created, the Rabbit will be powered down, and testing will stop.
# 
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
#     * i210  = v3.255
#     * BIOS  = v0.3.0
#    Rabbit-S:
#     * uBoot = v1.11
#     * sFPGA = v2.05
#     * PAX   = v4.90, bld B464
#    E3.s Kioxia Drive FW = ver 1TCRS104
#===================================================================

Passed_arg=$1
# echo -e "\n\n Passed argument is $Passed_arg"

# Make sure an arguement was passed . . .
if [ $# -eq 0 ]; then      
      echo -e "\n\n Need to include a Chassis number to be tested!"
      echo -e "\n Argument passed to this script must be a valid Chassis # (0-7),"
      echo -e "      tds (Chassis 1 & 3), or all (Rabbits in all 8 Chassis tested)!"
      echo -e "      EX1:   ./Frontend.sh 5   --> test Rabbit in Chassis #5"
      echo -e "      EX2:   ./Frontend.sh all --> test Rabbits in all 8 Chassis"
      echo -e "      EXITING TEST -- Please re-launch w/valid argument."
      exit -1
fi

case $Passed_arg in
   0)
      echo -e "\n Testing Chassis 0"
      nohup ./FWnCablingTest.sh 0 &>/dev/null &
      Chassis0_pid=$!
      ;;
   1)
      echo -e "\n Testing Chassis 1"
      nohup ./FWnCablingTest.sh 1 &>/dev/null &
      Chassis1_pid=$!
      ;;
   2)
      echo -e "\n Testing Chassis 2"
      nohup ./FWnCablingTest.sh 2 &>/dev/null &
      Chassis2_pid=$!
      ;;
   3)
      echo -e "\n Testing Chassis 3"
      nohup ./FWnCablingTest.sh 3 &>/dev/null &
      Chassis3_pid=$!
      ;;
   4)
      echo -e "\n Testing Chassis 4"
      nohup ./FWnCablingTest.sh 4 &>/dev/null &
      Chassis4_pid=$!
      ;;
   5)
      echo -e "\n Testing Chassis 5"
      nohup ./FWnCablingTest.sh 5 &>/dev/null &
      Chassis5_pid=$!
      ;;
   6)
      echo -e "\n Testing Chassis 6"
      nohup ./FWnCablingTest.sh 6 &>/dev/null &
      Chassis6_pid=$!
      ;;
   7)
      echo -e "\n Testing Chassis 7"
      nohup ./FWnCablingTest.sh 7 &>/dev/null &
      Chassis7_pid=$!
      ;;
   tds)
      echo -e "\n TDS - Testing Chassis 1 & 3"
      nohup ./FWnCablingTest.sh 1 &>/dev/null &
      Chassis1_pid=$!
      nohup ./FWnCablingTest.sh 3 &>/dev/null &
      Chassis3_pid=$!
      ;;
   all)
      echo -e "\n Testing all Chassis (0-7)"
      nohup ./FWnCablingTest.sh 0 &>/dev/null &
      Chassis0_pid=$!
      nohup ./FWnCablingTest.sh 1 &>/dev/null &
      Chassis1_pid=$!
      nohup ./FWnCablingTest.sh 2 &>/dev/null &
      Chassis2_pid=$!
      nohup ./FWnCablingTest.sh 3 &>/dev/null &
      Chassis3_pid=$!
      nohup ./FWnCablingTest.sh 4 &>/dev/null &
      Chassis4_pid=$!
      nohup ./FWnCablingTest.sh 5 &>/dev/null &
      Chassis5_pid=$!
      nohup ./FWnCablingTest.sh 6 &>/dev/null &
      Chassis6_pid=$!
      nohup ./FWnCablingTest.sh 7 &>/dev/null &
      Chassis7_pid=$!
      ;;
   *)
      echo -e "\n Argument passed to this script must be a valid Chassis # (0-7),"
      echo -e "      tds (Chassis 1 & 3), or all (Rabbits in all 8 Chassis tested)!"
      echo -e "      EX1:   ./Frontend.sh 5   --> test Rabbit in Chassis #5"
      echo -e "      EX2:   ./Frontend.sh all --> test Rabbits in all 8 Chassis"
      echo -e "      EXITING TEST -- Please re-launch w/valid argument."
      exit -1
      ;;
esac

sleep 15m

case $Passed_arg in
   0)
      echo -e "\n Been 15min, see if Chassis 0 done testing . . ."
      delay=60
         while [ $delay -gt 0 ] ;
         do
             ps -p $Chassis0_pid > /dev/null
             if [ $? -eq 0 ]; then
                echo "Job still in-progress, give it another minute, $delay"
                sleep 1m
             elif [ $delay == 1 ]; then
                  echo -e "\n\nTesting should have finished by now - we've got a problem!!!\n"
                  echo -e "Exiting test!!"
                  exit -1
             else
                  echo "Job process is all done!"
                  delay=1
             fi
         ((delay--))
         done
         FILE=$(pwd)/logs/Chassis0_Failure.log
         if [ -f "$FILE" ]; then
            echo "Chassis0 had a failure - check error log for details!"
            exit -1
         fi
      ;;
   1)
      echo -e "\n Been 15min, see if Chassis 1 done testing . . ."
      delay=60
         while [ $delay -gt 0 ] ;
         do
             ps -p $Chassis1_pid > /dev/null
             if [ $? -eq 0 ]; then
                echo "Job still in-progress, give it another minute, $delay"
                sleep 1m
             elif [ $delay == 1 ]; then
                  echo -e "\n\nTesting should have finished by now - we've got a problem!!!\n"
                  echo -e "Exiting test!!"
                  exit -1
             else
                  echo "Job process is all done!"
                  delay=1
             fi
         ((delay--))
         done
         FILE=$(pwd)/logs/Chassis1_Failure.log
         if [ -f "$FILE" ]; then
            echo "Chassis1 had a failure - check error log for details!"
            exit -1
         fi
      ;;
   2)
      echo -e "\n Been 15min, see if Chassis 2 done testing . . ."
      delay=60
         while [ $delay -gt 0 ] ;
         do
             ps -p $Chassis2_pid > /dev/null
             if [ $? -eq 0 ]; then
                echo "Job still in-progress, give it another minute, $delay"
                sleep 1m
             elif [ $delay == 1 ]; then
                  echo -e "\n\nTesting should have finished by now - we've got a problem!!!\n"
                  echo -e "Exiting test!!"
                  exit -1
             else
                  echo "Job process is all done!"
                  delay=1
             fi
         ((delay--))
         done
         FILE=$(pwd)/logs/Chassis2_Failure.log
         if [ -f "$FILE" ]; then
            echo "Chassis2 had a failure - check error log for details!"
            exit -1
         fi
      ;;
   3)
      echo -e "\n Been 15min, see if Chassis 3 done testing . . ."
      delay=60
         while [ $delay -gt 0 ] ;
         do
             ps -p $Chassis3_pid > /dev/null
             if [ $? -eq 0 ]; then
                echo "Job still in-progress, give it another minute, $delay"
                sleep 1m
             elif [ $delay == 1 ]; then
                  echo -e "\n\nTesting should have finished by now - we've got a problem!!!\n"
                  echo -e "Exiting test!!"
                  exit -1
             else
                  echo "Job process is all done!"
                  delay=1
             fi
         ((delay--))
         done
         FILE=$(pwd)/logs/Chassis3_Failure.log
         if [ -f "$FILE" ]; then
            echo "Chassis3 had a failure - check error log for details!"
            exit -1
         fi
      ;;
   4)
      echo -e "\n Been 15min, see if Chassis 4 done testing . . ."
      delay=60
         while [ $delay -gt 0 ] ;
         do
             ps -p $Chassis4_pid > /dev/null
             if [ $? -eq 0 ]; then
                echo "Job still in-progress, give it another minute, $delay"
                sleep 1m
             elif [ $delay == 1 ]; then
                  echo -e "\n\nTesting should have finished by now - we've got a problem!!!\n"
                  echo -e "Exiting test!!"
                  exit -1
             else
                  echo "Job process is all done!"
                  delay=1
             fi
         ((delay--))
         done
         FILE=$(pwd)/logs/Chassis4_Failure.log
         if [ -f "$FILE" ]; then
            echo "Chassis4 had a failure - check error log for details!"
            exit -1
         fi
      ;;
   5)
      echo -e "\n Been 15min, see if Chassis 5 done testing . . ."
      delay=60
         while [ $delay -gt 0 ] ;
         do
             ps -p $Chassis5_pid > /dev/null
             if [ $? -eq 0 ]; then
                echo "Job still in-progress, give it another minute, $delay"
                sleep 1m
             elif [ $delay == 1 ]; then
                  echo -e "\n\nTesting should have finished by now - we've got a problem!!!\n"
                  echo -e "Exiting test!!"
                  exit -1
             else
                  echo "Job process is all done!"
                  delay=1
             fi
         ((delay--))
         done
         FILE=$(pwd)/logs/Chassis5_Failure.log
         if [ -f "$FILE" ]; then
            echo "Chassis5 had a failure - check error log for details!"
            exit -1
         fi
      ;;
   6)
      echo -e "\n Been 15min, see if Chassis 6 done testing . . ."
      delay=60
         while [ $delay -gt 0 ] ;
         do
             ps -p $Chassis6_pid > /dev/null
             if [ $? -eq 0 ]; then
                echo "Job still in-progress, give it another minute, $delay"
                sleep 1m
             elif [ $delay == 1 ]; then
                  echo -e "\n\nTesting should have finished by now - we've got a problem!!!\n"
                  echo -e "Exiting test!!"
                  exit -1
             else
                  echo "Job process is all done!"
                  delay=1
             fi
         ((delay--))
         done
         FILE=$(pwd)/logs/Chassis6_Failure.log
         if [ -f "$FILE" ]; then
            echo "Chassis6 had a failure - check error log for details!"
            exit -1
         fi
      ;;
   7)
      echo -e "\n Been 15min, see if Chassis 7 done testing . . ."
      delay=60
         while [ $delay -gt 0 ] ;
         do
             ps -p $Chassis7_pid > /dev/null
             if [ $? -eq 0 ]; then
                echo "Job still in-progress, give it another minute, $delay"
                sleep 1m
             elif [ $delay == 1 ]; then
                  echo -e "\n\nTesting should have finished by now - we've got a problem!!!\n"
                  echo -e "Exiting test!!"
                  exit -1
             else
                  echo "Job process is all done!"
                  delay=1
             fi
         ((delay--))
         done
         FILE=$(pwd)/logs/Chassis7_Failure.log
         if [ -f "$FILE" ]; then
            echo "Chassis7 had a failure - check error log for details!"
            exit -1
         fi
      ;;
   tds)
      echo -e "\n Been 15min, see if Chassis 1 & 3 are done testing . . ."
      delay=60
         while [ $delay -gt 0 ] ;
         do
             ps -p $Chassis1_pid > /dev/null
             if [ $? -eq 0 ]; then
                echo "Job still in-progress, give it another minute, $delay"
                sleep 1m
             elif [ $delay == 1 ]; then
                  echo -e "\n\nTesting should have finished by now - we've got a problem!!!\n"
                  echo -e "Exiting test!!"
                  exit -1
             else
                  echo "Job process is all done!"
                  delay=1
             fi
         ((delay--))
         done
         FILE=$(pwd)/logs/Chassis1_Failure.log
         if [ -f "$FILE" ]; then
            echo "Chassis1 had a failure - check error log for details!"
            exit -1
         fi
      echo -e "Chassis1 done testing! Now lets check Chassis #3 . . ."
      delay=60
         while [ $delay -gt 0 ] ;
         do
             ps -p $Chassis3_pid > /dev/null
             if [ $? -eq 0 ]; then
                echo "Job still in-progress, give it another minute, $delay"
                sleep 1m
             elif [ $delay == 1 ]; then
                  echo -e "\n\nTesting should have finished by now - we've got a problem!!!\n"
                  echo -e "Exiting test!!"
                  exit -1
             else
                  echo "Job process is all done!"
                  delay=1
             fi
         ((delay--))
         done
         FILE=$(pwd)/logs/Chassis3_Failure.log
         if [ -f "$FILE" ]; then
            echo "Chassis3 had a failure - check error log for details!"
            exit -1
         fi
      ;;
   all)
      echo -e "\n Been 15min, see if the 8 Chassis are done testing . . ."
      delay=60
         while [ $delay -gt 0 ] ;
         do
             ps -p $Chassis0_pid > /dev/null
             if [ $? -eq 0 ]; then
                echo "Job still in-progress, give it another minute, $delay"
                sleep 1m
             elif [ $delay == 1 ]; then
                  echo -e "\n\nTesting should have finished by now - we've got a problem!!!\n"
                  echo -e "Exiting test!!"
                  exit -1
             else
                  echo "Job process is all done!"
                  delay=1
             fi
         ((delay--))
         done
         FILE=$(pwd)/logs/Chassis0_Failure.log
         if [ -f "$FILE" ]; then
            echo "Chassis0 had a failure - check error log for details!"
            exit -1
         fi
      echo -e "Chassis0 done testing! Now lets check Chassis #1 . . ."
      delay=60
         while [ $delay -gt 0 ] ;
         do
             ps -p $Chassis1_pid > /dev/null
             if [ $? -eq 0 ]; then
                echo "Job still in-progress, give it another minute, $delay"
                sleep 1m
             elif [ $delay == 1 ]; then
                  echo -e "\n\nTesting should have finished by now - we've got a problem!!!\n"
                  echo -e "Exiting test!!"
                  exit -1
             else
                  echo "Job process is all done!"
                  delay=1
             fi
         ((delay--))
         done
         FILE=$(pwd)/logs/Chassis1_Failure.log
         if [ -f "$FILE" ]; then
            echo "Chassis1 had a failure - check error log for details!"
            exit -1
         fi
      echo -e "Chassis1 done testing! Now lets check Chassis #2 . . ."
      delay=60
         while [ $delay -gt 0 ] ;
         do
             ps -p $Chassis2_pid > /dev/null
             if [ $? -eq 0 ]; then
                echo "Job still in-progress, give it another minute, $delay"
                sleep 1m
             elif [ $delay == 1 ]; then
                  echo -e "\n\nTesting should have finished by now - we've got a problem!!!\n"
                  echo -e "Exiting test!!"
                  exit -1
             else
                  echo "Job process is all done!"
                  delay=1
             fi
         ((delay--))
         done
         FILE=$(pwd)/logs/Chassis2_Failure.log
         if [ -f "$FILE" ]; then
            echo "Chassis2 had a failure - check error log for details!"
            exit -1
         fi
      echo -e "Chassis2 done testing! Now lets check Chassis #3 . . ."
      delay=60
         while [ $delay -gt 0 ] ;
         do
             ps -p $Chassis3_pid > /dev/null
             if [ $? -eq 0 ]; then
                echo "Job still in-progress, give it another minute, $delay"
                sleep 1m
             elif [ $delay == 1 ]; then
                  echo -e "\n\nTesting should have finished by now - we've got a problem!!!\n"
                  echo -e "Exiting test!!"
                  exit -1
             else
                  echo "Job process is all done!"
                  delay=1
             fi
         ((delay--))
         done
         FILE=$(pwd)/logs/Chassis3_Failure.log
         if [ -f "$FILE" ]; then
            echo "Chassis3 had a failure - check error log for details!"
            exit -1
         fi
      echo -e "Chassis3 done testing! Now lets check Chassis #4 . . ."
      delay=60
         while [ $delay -gt 0 ] ;
         do
             ps -p $Chassis4_pid > /dev/null
             if [ $? -eq 0 ]; then
                echo "Job still in-progress, give it another minute, $delay"
                sleep 1m
             elif [ $delay == 1 ]; then
                  echo -e "\n\nTesting should have finished by now - we've got a problem!!!\n"
                  echo -e "Exiting test!!"
                  exit -1
             else
                  echo "Job process is all done!"
                  delay=1
             fi
         ((delay--))
         done
         FILE=$(pwd)/logs/Chassis4_Failure.log
         if [ -f "$FILE" ]; then
            echo "Chassis4 had a failure - check error log for details!"
            exit -1
         fi
      echo -e "Chassis4 done testing! Now lets check Chassis #5 . . ."
      delay=60
         while [ $delay -gt 0 ] ;
         do
             ps -p $Chassis5_pid > /dev/null
             if [ $? -eq 0 ]; then
                echo "Job still in-progress, give it another minute, $delay"
                sleep 1m
             elif [ $delay == 1 ]; then
                  echo -e "\n\nTesting should have finished by now - we've got a problem!!!\n"
                  echo -e "Exiting test!!"
                  exit -1
             else
                  echo "Job process is all done!"
                  delay=1
             fi
         ((delay--))
         done
         FILE=$(pwd)/logs/Chassis5_Failure.log
         if [ -f "$FILE" ]; then
            echo "Chassis5 had a failure - check error log for details!"
            exit -1
         fi
      echo -e "Chassis5 done testing! Now lets check Chassis #6 . . ."
      delay=60
         while [ $delay -gt 0 ] ;
         do
             ps -p $Chassis6_pid > /dev/null
             if [ $? -eq 0 ]; then
                echo "Job still in-progress, give it another minute, $delay"
                sleep 1m
             elif [ $delay == 1 ]; then
                  echo -e "\n\nTesting should have finished by now - we've got a problem!!!\n"
                  echo -e "Exiting test!!"
                  exit -1
             else
                  echo "Job process is all done!"
                  delay=1
             fi
         ((delay--))
         done
         FILE=$(pwd)/logs/Chassis6_Failure.log
         if [ -f "$FILE" ]; then
            echo "Chassis6 had a failure - check error log for details!"
            exit -1
         fi
      echo -e "Chassis6 done testing! Now lets check Chassis #7 . . ."
      delay=60
         while [ $delay -gt 0 ] ;
         do
             ps -p $Chassis7_pid > /dev/null
             if [ $? -eq 0 ]; then
                echo "Job still in-progress, give it another minute, $delay"
                sleep 1m
             elif [ $delay == 1 ]; then
                  echo -e "\n\nTesting should have finished by now - we've got a problem!!!\n"
                  echo -e "Exiting test!!"
                  exit -1
             else
                  echo "Job process is all done!"
                  delay=1
             fi
         ((delay--))
         done
         FILE=$(pwd)/logs/Chassis7_Failure.log
         if [ -f "$FILE" ]; then
            echo "Chassis7 had a failure - check error log for details!"
            exit -1
         fi
      ;;
   *)
      echo -e "\n Something went really wacky if we got here . . ."
      echo -e "Exiting test, bye-bye!"
      exit -1
      ;;
esac

echo -e "\n\nAll testing complete - check for presence of Error logs in /logs/ directory."
