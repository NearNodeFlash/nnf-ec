#!/bin/bash
emmc-setup -B
cp nc-1.8.3-13.itb /boot/a.itb
emmc-setup -b
echo -n A > /dev/mmcblk0p1
reboot
