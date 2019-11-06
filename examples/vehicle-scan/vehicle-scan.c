/*
 *  vehicle-scan.c
 *
 *  drive-sdk - Anki Drive SDK
 *
 *  Copyright (C) 2014 Anki, Inc.
 *
 *  This program is free software; you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation; either version 2 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program; if not, write to the Free Software
 *  Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 *  Portions of this software are derived from BlueZ, a Bluetooth protocol stack for
 *  Linux. The license for BlueZ is included below.
 */

/*
 *
 *  BlueZ - Bluetooth protocol stack for Linux
 *
 *  Copyright (C) 2000-2001  Qualcomm Incorporated
 *  Copyright (C) 2002-2003  Maxim Krasnyansky <maxk@qualcomm.com>
 *  Copyright (C) 2002-2010  Marcel Holtmann <marcel@holtmann.org>
 *
 *
 *  This program is free software; you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation; either version 2 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program; if not, write to the Free Software
 *  Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 */

#include <stdio.h>
#include <errno.h>
#include <ctype.h>
#include <fcntl.h>
#include <unistd.h>
#include <stdlib.h>
#include <string.h>
#include <getopt.h>
#include <sys/param.h>
#include <sys/ioctl.h>
#include <sys/socket.h>
#include <signal.h>

#define for_each_opt(opt, long, short) while ((opt=getopt_long(argc, argv, short ? short:"+", long, NULL)) != -1)

#include <bzle/bluetooth/bluetooth.h>
#include <bzle/bluetooth/hci.h>
#include <bzle/bluetooth/hci_lib.h>

#include "uthash.h"

#include <ankidrive/advertisement.h>

/* Unofficial value, might still change */
#define LE_LINK         0x03

#define FLAGS_AD_TYPE 0x01
#define FLAGS_LIMITED_MODE_BIT 0x01
#define FLAGS_GENERAL_MODE_BIT 0x02

#define EIR_FLAGS                   0x01  /* flags */
#define EIR_UUID16_SOME             0x02  /* 16-bit UUID, more available */
#define EIR_UUID16_ALL              0x03  /* 16-bit UUID, all listed */
#define EIR_UUID32_SOME             0x04  /* 32-bit UUID, more available */
#define EIR_UUID32_ALL              0x05  /* 32-bit UUID, all listed */
#define EIR_UUID128_SOME            0x06  /* 128-bit UUID, more available */
#define EIR_UUID128_ALL             0x07  /* 128-bit UUID, all listed */
#define EIR_NAME_SHORT              0x08  /* shortened local name */
#define EIR_NAME_COMPLETE           0x09  /* complete local name */
#define EIR_TX_POWER                0x0A  /* transmit power level */
#define EIR_DEVICE_ID               0x10  /* device ID */

struct _vehicle {
    bdaddr_t   address; 
    anki_vehicle_adv_t adv;
    uint8_t scan_complete;
    UT_hash_handle hh;
};
typedef struct _vehicle vehicle_t;

static volatile int signal_received = 0;

static void hex_dump(char *pref, int width, unsigned char *buf, int len)
{
        register int i,n;

        for (i = 0, n = 1; i < len; i++, n++) {
                if (n == 1)
                        printf("%s", pref);
                printf("%2.2X ", buf[i]);
                if (n == width) {
                        printf("\n");
                        n = 0;
                }
        }
        if (i && n!=1)
                printf("\n");
}

static struct option lescan_options[] = {
        { "help",       0, 0, 'h' },
        { "privacy",    0, 0, 'p' },
        { "passive",    0, 0, 'P' },
        { "whitelist",  0, 0, 'w' },
        { "discovery",  1, 0, 'd' },
        { "duplicates", 0, 0, 'D' },
        { 0, 0, 0, 0 }
};

static const char *lescan_help =
        "Usage:\n"
        "\tlescan [--privacy] enable privacy\n"
        "\tlescan [--passive] set scan type passive (default active)\n"
        "\tlescan [--whitelist] scan for address in the whitelist only\n"
        "\tlescan [--discovery=g|l] enable general or limited discovery"
                "procedure\n"
        "\tlescan [--duplicates] don't filter duplicates\n";

static void helper_arg(int min_num_arg, int max_num_arg, int *argc,
                        char ***argv, const char *usage)
{
        *argc -= optind;
        /* too many arguments, but when "max_num_arg < min_num_arg" then no
                 limiting (prefer "max_num_arg=-1" to gen infinity)
        */
        if ( (*argc > max_num_arg) && (max_num_arg >= min_num_arg ) ) {
                fprintf(stderr, "%s: too many arguments (maximal: %i)\n",
                                *argv[0], max_num_arg);
                printf("%s", usage);
                exit(1);
        }

        /* print usage */
        if (*argc < min_num_arg) {
                fprintf(stderr, "%s: too few arguments (minimal: %i)\n",
                                *argv[0], min_num_arg);
                printf("%s", usage);
                exit(0);
        }

        *argv += optind;
}


static void sigint_handler(int sig)
{
        signal_received = sig;
}

static int read_flags(uint8_t *flags, const uint8_t *data, size_t size)
{
        size_t offset;

        if (!flags || !data)
                return -EINVAL;

        offset = 0;
        while (offset < size) {
                uint8_t len = data[offset];
                uint8_t type;

                /* Check if it is the end of the significant part */
                if (len == 0)
                        break;

                if (len + offset > size)
                        break;

                type = data[offset + 1];

                if (type == FLAGS_AD_TYPE) {
                        *flags = data[offset + 2];
                        return 0;
                }

                offset += 1 + len;
        }

        return -ENOENT;
}

static int check_report_filter(uint8_t procedure, le_advertising_info *info)
{
        uint8_t flags;

        /* If no discovery procedure is set, all reports are treat as valid */
        if (procedure == 0)
                return 1;

        /* Read flags AD type value from the advertising report if it exists */
        if (read_flags(&flags, info->data, info->length))
                return 0;

        switch (procedure) {
        case 'l': /* Limited Discovery Procedure */
                if (flags & FLAGS_LIMITED_MODE_BIT)
                        return 1;
                break;
        case 'g': /* General Discovery Procedure */
                if (flags & (FLAGS_LIMITED_MODE_BIT | FLAGS_GENERAL_MODE_BIT))
                        return 1;
                break;
        default:
                fprintf(stderr, "Unknown discovery procedure\n");
        }

        return 0;
}

static void eir_parse_name(uint8_t *eir, size_t eir_len,
                                                char *buf, size_t buf_len)
{
        size_t offset;

        offset = 0;
        while (offset < eir_len) {
                uint8_t field_len = eir[0];
                size_t name_len;

                /* Check for the end of EIR */
                if (field_len == 0)
                        break;

                if (offset + field_len > eir_len)
                        goto failed;

                printf("eir_type: %02x: ", eir[1]);
		hex_dump("0x", 32, &eir[2], field_len-1); 

                switch (eir[1]) {
                case EIR_NAME_SHORT:
                case EIR_NAME_COMPLETE:
                        name_len = field_len - 1;
                        if (name_len > buf_len)
                                goto failed;

                        memcpy(buf, &eir[2], name_len);
                        return;
                }

                offset += field_len + 1;
                eir += field_len + 1;
        }

failed:
        snprintf(buf, buf_len, "(unknown)");
}

vehicle_t *vehicles = NULL;

static int print_advertising_devices(int dd, uint8_t filter_type)
{
        unsigned char buf[HCI_MAX_EVENT_SIZE], *ptr;
        struct hci_filter nf, of;
        struct sigaction sa;
        socklen_t olen;
        int len;

        olen = sizeof(of);
        if (getsockopt(dd, SOL_HCI, HCI_FILTER, &of, &olen) < 0) {
                printf("Could not get socket options\n");
                return -1;
        }

        hci_filter_clear(&nf);
        hci_filter_set_ptype(HCI_EVENT_PKT, &nf);
        hci_filter_set_event(EVT_LE_META_EVENT, &nf);

        if (setsockopt(dd, SOL_HCI, HCI_FILTER, &nf, sizeof(nf)) < 0) {
                printf("Could not set socket options\n");
                return -1;
        }

        memset(&sa, 0, sizeof(sa));
        sa.sa_flags = SA_NOCLDSTOP;
        sa.sa_handler = sigint_handler;
        sigaction(SIGINT, &sa, NULL);

        while (1) {
                evt_le_meta_event *meta;
                le_advertising_info *info;
                char addr[18];

                while ((len = read(dd, buf, sizeof(buf))) < 0) {
                        if (errno == EINTR && signal_received == SIGINT) {
                                len = 0;
                                goto done;
                        }

                        if (errno == EAGAIN || errno == EINTR)
                                continue;
                        goto done;
                }

                ptr = buf + (1 + HCI_EVENT_HDR_SIZE);
                len -= (1 + HCI_EVENT_HDR_SIZE);

                meta = (void *) ptr;

                if (meta->subevent != 0x02)
                        goto done;

                /* Ignoring multiple reports */
                info = (le_advertising_info *) (meta->data + 1);
                if (check_report_filter(filter_type, info)) {
                        char name[30];

                        memset(name, 0, sizeof(name));

                        ba2str(&info->bdaddr, addr);

                        vehicle_t vehicle;
                        vehicle_t *v = &vehicle;

                        HASH_FIND(hh, vehicles, &info->bdaddr, sizeof(bdaddr_t), v);
                        if (v == NULL) {
                            v = (vehicle_t *)malloc(sizeof(vehicle_t));
                            memset(&v->adv, 0, sizeof(anki_vehicle_adv_t));
                            memcpy(&v->address, &info->bdaddr, sizeof(bdaddr_t));
                            HASH_ADD(hh, vehicles, address, sizeof(bdaddr_t), v);
                        }
                       
                        int err = anki_vehicle_parse_adv_record(info->data, info->length, &v->adv);
                        
                        if (err == 0 && v->adv.mfg_data.identifier > 0 && v->adv.local_name.version > 0 && !v->scan_complete) {
                            v->scan_complete = 1;
                            printf("%s ", addr);

                            const char *model = NULL;
                            switch(v->adv.mfg_data.model_id) {
                                case 1:
                                    model = "Kourai";
                                    break;
                                case 2:
                                    model = "Boson";
                                    break;
                                case 3:
                                    model = "Rho";
                                    break;
                                case 4:
                                    model = "Katal";
                                    break;
                                case 8:
                                    model = "GroundShock";
                                    break;
                                case 9:
                                    model = "Skull";
                                    break;
                                default:
                                    model = "Unknown";
                                    break;
                            }

                            printf("%s [v%04x] (%d=%s %04x)\n", v->adv.local_name.name, v->adv.local_name.version & 0xffff, v->adv.mfg_data.model_id, model, v->adv.mfg_data.identifier & 0xffff);
                        }
                }
        }

done:
        setsockopt(dd, SOL_HCI, HCI_FILTER, &of, sizeof(of));

        if (len < 0)
                return -1;

        return 0;
}

static void cmd_lescan(int dev_id, int argc, char **argv)
{
        int err, opt, dd;
        uint8_t own_type = 0x00;
        uint8_t scan_type = 0x01;
        uint8_t filter_type = 0;
        uint8_t filter_policy = 0x00;
        uint16_t interval = htobs(0x0010);
        uint16_t window = htobs(0x0010);
        uint8_t filter_dup = 1;

        for_each_opt(opt, lescan_options, NULL) {
                switch (opt) {
                case 'p':
                        own_type = 0x01; /* Random */
                        break;
                case 'P':
                        scan_type = 0x00; /* Passive */
                        break;
                case 'w':
                        filter_policy = 0x01; /* Whitelist */
                        break;
                case 'd':
                        filter_type = optarg[0];
                        if (filter_type != 'g' && filter_type != 'l') {
                                fprintf(stderr, "Unknown discovery procedure\n");
                                exit(1);
                        }

                        interval = htobs(0x0012);
                        window = htobs(0x0012);
                        break;
                case 'D':
                        filter_dup = 0x00;
                        break;
                default:
                        printf("%s", lescan_help);
                        return;
                }
        }
        helper_arg(0, 1, &argc, &argv, lescan_help);

        if (dev_id < 0)
                dev_id = hci_get_route(NULL);

        dd = hci_open_dev(dev_id);
        if (dd < 0) {
                perror("Could not open device");
                exit(1);
                exit(1);
        }

        err = hci_le_set_scan_parameters(dd, scan_type, interval, window,
                                                own_type, filter_policy, 2000);
        if (err < 0) {
                perror("Set scan parameters failed");
                exit(1);
        }

        err = hci_le_set_scan_enable(dd, 0x01, filter_dup, 2000);
        if (err < 0) {
                perror("Enable scan failed");
                exit(1);
        }

        printf("LE Scan ...\n");

        err = print_advertising_devices(dd, filter_type);
        if (err < 0) {
                perror("Could not receive advertising events");
                exit(1);
        }

        err = hci_le_set_scan_enable(dd, 0x00, filter_dup, 2000);
        if (err < 0) {
                fprintf(stderr, "Disable scan failed with error code: %d", err);
                hci_close_dev(dd);
                exit(1);
        }

        hci_close_dev(dd);
}

int main(int argc, char *argv[]) {
        int opt, i, dev_id = -1;

	cmd_lescan(dev_id, argc, argv);

	return 0;
}
