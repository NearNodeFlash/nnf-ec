# Tools for low-level device manipulation

The scripts in this directory are meant for low-level NVMe device manipulation as well as Microchip Pax switch debugging across. For individual script details, execute the script with the `-h` option.

**NOTE**: All of the following scripts iterate through the KIOXIA drives in the system and perform the same operation on each.

## NVMe Drive access

`bind.sh` sets up PCIe access to each drive connected to the Microchip PAX switches. Run this script **once** after a PAX powercycle to access the drives. Once the drives are bound, use the `nvme.sh` script to send commands to the drives.

## Creating an LVM volume

To create an LVM volume, the following sequence of steps is required:

- Create a namespace on each drive
- Attach that namespace to a controller
- Create the LVM volume from those namespaces

```bash
[root@RabbitP-CPU tools]# ./nvme.sh list
Namespaces on 0x1600@/dev/switchtec0
Namespaces on 0x1900@/dev/switchtec0
Namespaces on 0x1700@/dev/switchtec0
Namespaces on 0x1400@/dev/switchtec0
Namespaces on 0x1800@/dev/switchtec0
Namespaces on 0x1a00@/dev/switchtec0
Namespaces on 0x1300@/dev/switchtec0
Namespaces on 0x1500@/dev/switchtec0
Namespaces on 0x4100@/dev/switchtec1
Namespaces on 0x3e00@/dev/switchtec1
Namespaces on 0x3d00@/dev/switchtec1
Namespaces on 0x3c00@/dev/switchtec1
Namespaces on 0x4000@/dev/switchtec1
Namespaces on 0x4200@/dev/switchtec1
Namespaces on 0x3b00@/dev/switchtec1
Namespaces on 0x3f00@/dev/switchtec1
[root@RabbitP-CPU tools]# ./nvme.sh create
Reading Capacity on 0x1600@/dev/switchtec0
Creating Namespaces on 0x1600@/dev/switchtec0 with size 3840755982336
create-ns: Success, created nsid:1
Reading Capacity on 0x1900@/dev/switchtec0
Creating Namespaces on 0x1900@/dev/switchtec0 with size 3840755982336
create-ns: Success, created nsid:1
Reading Capacity on 0x1700@/dev/switchtec0
Creating Namespaces on 0x1700@/dev/switchtec0 with size 3840755982336
create-ns: Success, created nsid:1
Reading Capacity on 0x1400@/dev/switchtec0
Creating Namespaces on 0x1400@/dev/switchtec0 with size 3840755982336
create-ns: Success, created nsid:1
Reading Capacity on 0x1800@/dev/switchtec0
Creating Namespaces on 0x1800@/dev/switchtec0 with size 3840755982336
create-ns: Success, created nsid:1
Reading Capacity on 0x1a00@/dev/switchtec0
Creating Namespaces on 0x1a00@/dev/switchtec0 with size 3840755982336
create-ns: Success, created nsid:1
Reading Capacity on 0x1300@/dev/switchtec0
Creating Namespaces on 0x1300@/dev/switchtec0 with size 3840755982336
create-ns: Success, created nsid:1
Reading Capacity on 0x1500@/dev/switchtec0
Creating Namespaces on 0x1500@/dev/switchtec0 with size 3840755982336
create-ns: Success, created nsid:1
Reading Capacity on 0x4100@/dev/switchtec1
Creating Namespaces on 0x4100@/dev/switchtec1 with size 3840755982336
create-ns: Success, created nsid:1
Reading Capacity on 0x3e00@/dev/switchtec1
Creating Namespaces on 0x3e00@/dev/switchtec1 with size 3840755982336
create-ns: Success, created nsid:1
Reading Capacity on 0x3d00@/dev/switchtec1
Creating Namespaces on 0x3d00@/dev/switchtec1 with size 3840755982336
create-ns: Success, created nsid:1
Reading Capacity on 0x3c00@/dev/switchtec1
Creating Namespaces on 0x3c00@/dev/switchtec1 with size 3840755982336
create-ns: Success, created nsid:1
Reading Capacity on 0x4000@/dev/switchtec1
Creating Namespaces on 0x4000@/dev/switchtec1 with size 3840755982336
create-ns: Success, created nsid:1
Reading Capacity on 0x4200@/dev/switchtec1
Creating Namespaces on 0x4200@/dev/switchtec1 with size 1920383410176
create-ns: Success, created nsid:1
Reading Capacity on 0x3b00@/dev/switchtec1
Creating Namespaces on 0x3b00@/dev/switchtec1 with size 3840755982336
create-ns: Success, created nsid:1
Reading Capacity on 0x3f00@/dev/switchtec1
Creating Namespaces on 0x3f00@/dev/switchtec1 with size 3840755982336
create-ns: Success, created nsid:1
[root@RabbitP-CPU tools]# ./nvme.sh list
Namespaces on 0x1600@/dev/switchtec0
[   0]:0x1
Namespaces on 0x1900@/dev/switchtec0
[   0]:0x1
Namespaces on 0x1700@/dev/switchtec0
[   0]:0x1
Namespaces on 0x1400@/dev/switchtec0
[   0]:0x1
Namespaces on 0x1800@/dev/switchtec0
[   0]:0x1
Namespaces on 0x1a00@/dev/switchtec0
[   0]:0x1
Namespaces on 0x1300@/dev/switchtec0
[   0]:0x1
Namespaces on 0x1500@/dev/switchtec0
[   0]:0x1
Namespaces on 0x4100@/dev/switchtec1
[   0]:0x1
Namespaces on 0x3e00@/dev/switchtec1
[   0]:0x1
Namespaces on 0x3d00@/dev/switchtec1
[   0]:0x1
Namespaces on 0x3c00@/dev/switchtec1
[   0]:0x1
Namespaces on 0x4000@/dev/switchtec1
[   0]:0x1
Namespaces on 0x4200@/dev/switchtec1
[   0]:0x1
Namespaces on 0x3b00@/dev/switchtec1
[   0]:0x1
Namespaces on 0x3f00@/dev/switchtec1
[   0]:0x1
[root@RabbitP-CPU tools]# nvme list
Node                  SN                   Model                                    Namespace Usage                      Format           FW Rev
--------------------- -------------------- ---------------------------------------- --------- -------------------------- ---------------- --------
/dev/nvme0n1          S666NE0R800568       SAMSUNG MZ1L21T9HCLS-00A07               1         248.70  GB /   1.92  TB    512   B +  0 B   GDC7102Q
[root@RabbitP-CPU tools]# ./nvme.sh attach
Attaching Namespace 1 on 0x1600@/dev/switchtec0 to Controller 3
attach-ns: Success, nsid:1
Attaching Namespace 1 on 0x1900@/dev/switchtec0 to Controller 3
attach-ns: Success, nsid:1
Attaching Namespace 1 on 0x1700@/dev/switchtec0 to Controller 3
attach-ns: Success, nsid:1
Attaching Namespace 1 on 0x1400@/dev/switchtec0 to Controller 3
attach-ns: Success, nsid:1
Attaching Namespace 1 on 0x1800@/dev/switchtec0 to Controller 3
attach-ns: Success, nsid:1
Attaching Namespace 1 on 0x1a00@/dev/switchtec0 to Controller 3
attach-ns: Success, nsid:1
Attaching Namespace 1 on 0x1300@/dev/switchtec0 to Controller 3
attach-ns: Success, nsid:1
Attaching Namespace 1 on 0x1500@/dev/switchtec0 to Controller 3
attach-ns: Success, nsid:1
Attaching Namespace 1 on 0x4100@/dev/switchtec1 to Controller 3
attach-ns: Success, nsid:1
Attaching Namespace 1 on 0x3e00@/dev/switchtec1 to Controller 3
attach-ns: Success, nsid:1
Attaching Namespace 1 on 0x3d00@/dev/switchtec1 to Controller 3
attach-ns: Success, nsid:1
Attaching Namespace 1 on 0x3c00@/dev/switchtec1 to Controller 3
attach-ns: Success, nsid:1
Attaching Namespace 1 on 0x4000@/dev/switchtec1 to Controller 3
attach-ns: Success, nsid:1
Attaching Namespace 1 on 0x4200@/dev/switchtec1 to Controller 3
attach-ns: Success, nsid:1
Attaching Namespace 1 on 0x3b00@/dev/switchtec1 to Controller 3
attach-ns: Success, nsid:1
Attaching Namespace 1 on 0x3f00@/dev/switchtec1 to Controller 3
attach-ns: Success, nsid:1
[root@RabbitP-CPU tools]# nvme list
Node                  SN                   Model                                    Namespace Usage                      Format           FW Rev
--------------------- -------------------- ---------------------------------------- --------- -------------------------- ---------------- --------
/dev/nvme0n1          S666NE0R800568       SAMSUNG MZ1L21T9HCLS-00A07               1         248.70  GB /   1.92  TB    512   B +  0 B   GDC7102Q
/dev/nvme1n1          8C40A00W01TZ         KIOXIA KCM71RJE3T84                      1           0.00   B /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme10n1         8C40A00D01TZ         KIOXIA KCM71RJE3T84                      1           0.00   B /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme11n1         8C40A00R01TZ         KIOXIA KCM71RJE3T84                      1           0.00   B /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme12n1         8C40A00801TZ         KIOXIA KCM71RJE3T84                      1           0.00   B /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme13n1         8C40A00F01TZ         KIOXIA KCM71RJE3T84                      1           0.00   B /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme14n1         8C40A00C01TZ         KIOXIA KCM71RJE3T84                      1           0.00   B /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme15n1         8C40A00E01TZ         KIOXIA KCM71RJE3T84                      1           0.00   B /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme16n1         8C40A00J01TZ         KIOXIA KCM71RJE3T84                      1           0.00   B /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme2n1          8C40A00N01TZ         KIOXIA KCM71RJE3T84                      1           0.00   B /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme3n1          8C40A00K01TZ         KIOXIA KCM71RJE3T84                      1           0.00   B /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme4n1          8C40A00G01TZ         KIOXIA KCM71RJE3T84                      1           0.00   B /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme5n1          8C40A00901TZ         KIOXIA KCM71RJE3T84                      1           0.00   B /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme6n1          8CN0A01301SZ         KIOXIA KCM71RJE1T92                      1           0.00   B /   1.92  TB      4 KiB +  0 B   01A1
/dev/nvme7n1          8C40A00P01TZ         KIOXIA KCM71RJE3T84                      1           0.00   B /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme8n1          8C40A00M01TZ         KIOXIA KCM71RJE3T84                      1           0.00   B /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme9n1          8C40A00U01TZ         KIOXIA KCM71RJE3T84                      1           0.00   B /   3.84  TB      4 KiB +  0 B   01A1
[root@RabbitP-CPU tools]# ls /dev/nvme*n1
/dev/nvme0n1  /dev/nvme10n1  /dev/nvme11n1  /dev/nvme12n1  /dev/nvme13n1  /dev/nvme14n1  /dev/nvme15n1  /dev/nvme16n1  /dev/nvme1n1  /dev/nvme2n1  /dev/nvme3n1  /dev/nvme4n1  /dev/nvme5n1  /dev/nvme6n1  /dev/nvme7n1  /dev/nvme8n1  /dev/nvme9n1
[root@RabbitP-CPU tools]# ./lvm.sh create
  Found Kioxia drive /dev/nvme10n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme11n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme12n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme13n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme14n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme15n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme16n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme1n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme2n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme3n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme4n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme5n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme6n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme7n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme8n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme9n1
    Found Namespace 1
16 DRIVES: /dev/nvme10n1 /dev/nvme11n1 /dev/nvme12n1 /dev/nvme13n1 /dev/nvme14n1 /dev/nvme15n1 /dev/nvme16n1 /dev/nvme1n1 /dev/nvme2n1 /dev/nvme3n1 /dev/nvme4n1 /dev/nvme5n1 /dev/nvme6n1 /dev/nvme7n1 /dev/nvme8n1 /dev/nvme9n1
Creating Physical Volume '/dev/nvme10n1'
  Physical volume "/dev/nvme10n1" successfully created.
Creating Physical Volume '/dev/nvme11n1'
  Physical volume "/dev/nvme11n1" successfully created.
Creating Physical Volume '/dev/nvme12n1'
  Physical volume "/dev/nvme12n1" successfully created.
Creating Physical Volume '/dev/nvme13n1'
  Physical volume "/dev/nvme13n1" successfully created.
Creating Physical Volume '/dev/nvme14n1'
  Physical volume "/dev/nvme14n1" successfully created.
Creating Physical Volume '/dev/nvme15n1'
  Physical volume "/dev/nvme15n1" successfully created.
Creating Physical Volume '/dev/nvme16n1'
  Physical volume "/dev/nvme16n1" successfully created.
Creating Physical Volume '/dev/nvme1n1'
  Physical volume "/dev/nvme1n1" successfully created.
Creating Physical Volume '/dev/nvme2n1'
  Physical volume "/dev/nvme2n1" successfully created.
Creating Physical Volume '/dev/nvme3n1'
  Physical volume "/dev/nvme3n1" successfully created.
Creating Physical Volume '/dev/nvme4n1'
  Physical volume "/dev/nvme4n1" successfully created.
Creating Physical Volume '/dev/nvme5n1'
  Physical volume "/dev/nvme5n1" successfully created.
Creating Physical Volume '/dev/nvme6n1'
  Physical volume "/dev/nvme6n1" successfully created.
Creating Physical Volume '/dev/nvme7n1'
  Physical volume "/dev/nvme7n1" successfully created.
Creating Physical Volume '/dev/nvme8n1'
  Physical volume "/dev/nvme8n1" successfully created.
Creating Physical Volume '/dev/nvme9n1'
  Physical volume "/dev/nvme9n1" successfully created.
Creating Volume Group 'rabbit'
  Volume group "rabbit" successfully created
Creating Logical Volume 'rabbit'
  Rounding size (14193459 extents) down to stripe boundary size (14193456 extents)
  WARNING: Logical volume rabbit/rabbit not zeroed.
  Logical volume "rabbit" created.
Activate Volume Group 'rabbit'
  1 logical volume(s) in volume group "rabbit" now active
DONE! Access the volume at /dev/rabbit/rabbit
[root@RabbitP-CPU tools]# lvs
  LV     VG                      Attr       LSize   Pool Origin Data%  Meta%  Move Log Cpy%Sync Convert
  rabbit rabbit                  -wi-a----- <27.95t
  home   toss_dhcp-10-30-105-181 -wi-ao---- 145.71g
  root   toss_dhcp-10-30-105-181 -wi-ao----  70.00g
  swap   toss_dhcp-10-30-105-181 -wi-ao---- <15.59g
[root@RabbitP-CPU tools]# vgs
  VG                      #PV #LV #SN Attr   VSize    VFree
  rabbit                   16   1   0 wz--n-   54.14t <26.20t
  toss_dhcp-10-30-105-181   1   3   0 wz--n- <231.30g      0
```

## Deleting an LVM volume

To delete an LVM volume, a semi-reversed set of steps is required:

- Delete the LVM volume
- Delete the namespaces on each drive (no need to `detach` those namespaces)

```bash
[root@RabbitP-CPU tools]# ./lvm.sh delete
Removing Logical Volume 'rabbit'
  Volume group name "{rabbit}" has invalid characters.
  Cannot process volume group {rabbit}
Deactivate Volume Group 'rabbit'
  0 logical volume(s) in volume group "rabbit" now active
Removing Volume Group'rabbit'
  Logical volume "rabbit" successfully removed.
  Volume group "rabbit" successfully removed
  Found Kioxia drive /dev/nvme10n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme11n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme12n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme13n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme14n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme15n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme16n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme1n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme2n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme3n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme4n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme5n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme6n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme7n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme8n1
    Found Namespace 1
  Found Kioxia drive /dev/nvme9n1
    Found Namespace 1
16 DRIVES: /dev/nvme10n1 /dev/nvme11n1 /dev/nvme12n1 /dev/nvme13n1 /dev/nvme14n1 /dev/nvme15n1 /dev/nvme16n1 /dev/nvme1n1 /dev/nvme2n1 /dev/nvme3n1 /dev/nvme4n1 /dev/nvme5n1 /dev/nvme6n1 /dev/nvme7n1 /dev/nvme8n1 /dev/nvme9n1
Remove Physical Volume '/dev/nvme10n1'
  Labels on physical volume "/dev/nvme10n1" successfully wiped.
[root@RabbitP-CPU tools]# lvs
  LV   VG                      Attr       LSize   Pool Origin Data%  Meta%  Move Log Cpy%Sync Convert
  home toss_dhcp-10-30-105-181 -wi-ao---- 145.71g
  root toss_dhcp-10-30-105-181 -wi-ao----  70.00g
  swap toss_dhcp-10-30-105-181 -wi-ao---- <15.59g
[root@RabbitP-CPU tools]# vgs
  VG                      #PV #LV #SN Attr   VSize    VFree
  toss_dhcp-10-30-105-181   1   3   0 wz--n- <231.30g    0
[root@RabbitP-CPU tools]# nvme list
Node                  SN                   Model                                    Namespace Usage                      Format           FW Rev
--------------------- -------------------- ---------------------------------------- --------- -------------------------- ---------------- --------
/dev/nvme0n1          S666NE0R800568       SAMSUNG MZ1L21T9HCLS-00A07               1         248.70  GB /   1.92  TB    512   B +  0 B   GDC7102Q
/dev/nvme1n1          8C40A00W01TZ         KIOXIA KCM71RJE3T84                      1         131.07  kB /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme10n1         8C40A00D01TZ         KIOXIA KCM71RJE3T84                      1         131.07  kB /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme11n1         8C40A00R01TZ         KIOXIA KCM71RJE3T84                      1         131.07  kB /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme12n1         8C40A00801TZ         KIOXIA KCM71RJE3T84                      1         131.07  kB /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme13n1         8C40A00F01TZ         KIOXIA KCM71RJE3T84                      1         131.07  kB /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme14n1         8C40A00C01TZ         KIOXIA KCM71RJE3T84                      1         131.07  kB /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme15n1         8C40A00E01TZ         KIOXIA KCM71RJE3T84                      1         131.07  kB /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme16n1         8C40A00J01TZ         KIOXIA KCM71RJE3T84                      1         131.07  kB /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme2n1          8C40A00N01TZ         KIOXIA KCM71RJE3T84                      1         131.07  kB /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme3n1          8C40A00K01TZ         KIOXIA KCM71RJE3T84                      1         131.07  kB /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme4n1          8C40A00G01TZ         KIOXIA KCM71RJE3T84                      1         131.07  kB /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme5n1          8C40A00901TZ         KIOXIA KCM71RJE3T84                      1         131.07  kB /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme6n1          8CN0A01301SZ         KIOXIA KCM71RJE1T92                      1         131.07  kB /   1.92  TB      4 KiB +  0 B   01A1
/dev/nvme7n1          8C40A00P01TZ         KIOXIA KCM71RJE3T84                      1         131.07  kB /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme8n1          8C40A00M01TZ         KIOXIA KCM71RJE3T84                      1         131.07  kB /   3.84  TB      4 KiB +  0 B   01A1
/dev/nvme9n1          8C40A00U01TZ         KIOXIA KCM71RJE3T84                      1         131.07  kB /   3.84  TB      4 KiB +  0 B   01A1
[root@RabbitP-CPU tools]# ./nvme.sh delete
Deleting Namespaces 1 on 0x1600@/dev/switchtec0
delete-ns: Success, deleted nsid:1
Deleting Namespaces 1 on 0x1900@/dev/switchtec0
delete-ns: Success, deleted nsid:1
Deleting Namespaces 1 on 0x1700@/dev/switchtec0
delete-ns: Success, deleted nsid:1
Deleting Namespaces 1 on 0x1400@/dev/switchtec0
delete-ns: Success, deleted nsid:1
Deleting Namespaces 1 on 0x1800@/dev/switchtec0
delete-ns: Success, deleted nsid:1
Deleting Namespaces 1 on 0x1a00@/dev/switchtec0
delete-ns: Success, deleted nsid:1
Deleting Namespaces 1 on 0x1300@/dev/switchtec0
delete-ns: Success, deleted nsid:1
Deleting Namespaces 1 on 0x1500@/dev/switchtec0
delete-ns: Success, deleted nsid:1
Deleting Namespaces 1 on 0x4100@/dev/switchtec1
delete-ns: Success, deleted nsid:1
Deleting Namespaces 1 on 0x3e00@/dev/switchtec1
delete-ns: Success, deleted nsid:1
Deleting Namespaces 1 on 0x3d00@/dev/switchtec1
delete-ns: Success, deleted nsid:1
Deleting Namespaces 1 on 0x3c00@/dev/switchtec1
delete-ns: Success, deleted nsid:1
Deleting Namespaces 1 on 0x4000@/dev/switchtec1
delete-ns: Success, deleted nsid:1
Deleting Namespaces 1 on 0x4200@/dev/switchtec1
delete-ns: Success, deleted nsid:1
Deleting Namespaces 1 on 0x3b00@/dev/switchtec1
delete-ns: Success, deleted nsid:1
Deleting Namespaces 1 on 0x3f00@/dev/switchtec1
delete-ns: Success, deleted nsid:1
[root@RabbitP-CPU tools]#
```

## Example script to walk the cycle

```bash
create-write-delete-namespaces.sh*
```
