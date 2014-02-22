/*
 * Copyright (c) 2014 Anki, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "greatest.h"
#include "anki_greatest.h"

#include "protocol.h"
#include "anki_util.h"

SUITE(vehicle_protocol);

TEST test_struct_attribute_packed(void) {
    ASSERT_EQ(sizeof(anki_vehicle_msg_sdk_mode_t), ANKI_VEHICLE_MSG_SDK_MODE_SIZE+ANKI_VEHICLE_MSG_BASE_SIZE);
    PASS();
}

TEST test_set_sdk_mode(void) {
    anki_vehicle_msg_t msg;
    uint8_t size = anki_vehicle_msg_set_sdk_mode(&msg, 1);
    ASSERT_EQ(size, 3);

    uint8_t expect[3] = { 0x02, 0x90, 0x01 };
    ASSERT_BYTES_EQ(&msg, expect, 3);
    PASS();
}

TEST test_set_speed(void) {
    anki_vehicle_msg_t msg;
    uint8_t size = anki_vehicle_msg_set_speed(&msg, 1000, 25000);
    ASSERT_EQ(size, sizeof(anki_vehicle_msg_set_speed_t));

    uint8_t expect[7] = { 0x06, 0x24, 0xe8, 0x03, 0xa8, 0x61, 0x00 };
    ASSERT_BYTES_EQ(&msg, expect, 7);
    PASS();
}

TEST test_set_offset_from_center(void) {
    anki_vehicle_msg_t msg;
    memset(&msg, 0, sizeof(anki_vehicle_msg_t));
    uint8_t size = anki_vehicle_msg_set_offset_from_road_center(&msg, 0.0);

    ASSERT_EQ(size, sizeof(anki_vehicle_msg_set_offset_from_road_center_t));

    uint8_t expect[6] = { 0x5, 0x2c, 0, 0, 0, 0 };
    ASSERT_BYTES_EQ(&msg, expect, 6);
    PASS();
}

TEST test_disconnect(void) {
    anki_vehicle_msg_t msg;
    anki_vehicle_msg_disconnect(&msg);

    uint8_t expect[2] = { ANKI_VEHICLE_MSG_BASE_SIZE, ANKI_VEHICLE_MSG_C2V_DISCONNECT };
    ASSERT_BYTES_EQ(&msg, expect, 2);

    PASS();
}

TEST test_get_version(void) {
    anki_vehicle_msg_t msg;
    uint8_t size = anki_vehicle_msg_get_version(&msg);

    ASSERT_EQ(size, 2);

    uint8_t expect[2] = { ANKI_VEHICLE_MSG_BASE_SIZE, ANKI_VEHICLE_MSG_C2V_VERSION_REQUEST };
    ASSERT_BYTES_EQ(&msg, expect, sizeof(expect));

    PASS();
}

TEST test_get_battery_level(void) {
    anki_vehicle_msg_t msg;
    uint8_t size = anki_vehicle_msg_get_battery_level(&msg);

    ASSERT_EQ(size, 2);

    uint8_t expect[2] = { ANKI_VEHICLE_MSG_BASE_SIZE, ANKI_VEHICLE_MSG_C2V_BATTERY_LEVEL_REQUEST };
    ASSERT_BYTES_EQ(&msg, expect, sizeof(expect));

    PASS();
}

GREATEST_SUITE(vehicle_protocol) {
    RUN_TEST(test_struct_attribute_packed);
    RUN_TEST(test_set_sdk_mode);
    RUN_TEST(test_set_speed);
    RUN_TEST(test_set_offset_from_center);
    RUN_TEST(test_disconnect);
    RUN_TEST(test_get_version);
    RUN_TEST(test_get_battery_level);
}
