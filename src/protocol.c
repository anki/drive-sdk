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
#include <assert.h>
#include <string.h>

#include "protocol.h"

#define ANKI_VEHICLE_MSG_TYPE_SIZE  2

uint8_t anki_vehicle_msg_set_sdk_mode(anki_vehicle_msg_t *message, uint8_t on)
{
    assert(message != NULL);

    anki_vehicle_msg_sdk_mode_t *msg = (anki_vehicle_msg_sdk_mode_t *)message;
    memset(msg, 0, sizeof(anki_vehicle_msg_sdk_mode_t));
    msg->size = ANKI_VEHICLE_MSG_SDK_MODE_SIZE;
    msg->msg_id = ANKI_VEHICLE_MSG_C2V_SDK_MODE;
    msg->on = on;

    return sizeof(anki_vehicle_msg_sdk_mode_t);
}

uint8_t anki_vehicle_msg_set_speed(anki_vehicle_msg_t *message, uint16_t speed_mm_per_sec, uint16_t accel_mm_per_sec2)
{
    assert(message != NULL);

    anki_vehicle_msg_set_speed_t *msg = (anki_vehicle_msg_set_speed_t *)message;
    memset(msg, 0, sizeof(anki_vehicle_msg_set_speed_t));
    msg->size = ANKI_VEHICLE_MSG_C2V_SET_SPEED_SIZE;
    msg->msg_id = ANKI_VEHICLE_MSG_C2V_SET_SPEED;
    msg->speed_mm_per_sec = speed_mm_per_sec;
    msg->accel_mm_per_sec2 = accel_mm_per_sec2;

    return sizeof(anki_vehicle_msg_set_speed_t);
} 

uint8_t anki_vehicle_msg_set_offset_from_road_center(anki_vehicle_msg_t *msg, float offset_mm)
{
    assert(msg != NULL);

    anki_vehicle_msg_set_offset_from_road_center_t *m = (anki_vehicle_msg_set_offset_from_road_center_t *)msg;
    memset(m, 0, sizeof(anki_vehicle_msg_set_offset_from_road_center_t));
    m->size = sizeof(anki_vehicle_msg_set_offset_from_road_center_t) - ANKI_VEHICLE_MSG_BASE_SIZE;
    m->msg_id = ANKI_VEHICLE_MSG_C2V_SET_OFFSET_FROM_ROAD_CENTER;
    m->offset_mm = offset_mm;

    return sizeof(anki_vehicle_msg_set_offset_from_road_center_t);
}

uint8_t anki_vehicle_msg_change_lane(anki_vehicle_msg_t *message, uint16_t horizontal_speed_mm_per_sec, float offset_from_center_mm)
{
    assert(message != NULL);
    memset(message, 0, sizeof(anki_vehicle_msg_t));

    anki_vehicle_msg_change_lane_t *msg = (anki_vehicle_msg_change_lane_t *)message;
    msg->size = ANKI_VEHICLE_MSG_C2V_CHANGE_LANE_SIZE;
    msg->msg_id = ANKI_VEHICLE_MSG_C2V_CHANGE_LANE;
    msg->horizontal_speed_mm_per_sec = horizontal_speed_mm_per_sec;
    msg->offset_from_road_center_mm = offset_from_center_mm;

    return sizeof(anki_vehicle_msg_change_lane_t);
}

uint8_t anki_vehicle_msg_set_lights(anki_vehicle_msg_t *message, uint8_t mask)
{
    assert(message != NULL);

    anki_vehicle_msg_set_lights_t *msg = (anki_vehicle_msg_set_lights_t *)message;
    memset(msg, 0, sizeof(anki_vehicle_msg_set_lights_t));
    msg->size = ANKI_VEHICLE_MSG_C2V_SET_LIGHTS_SIZE;
    msg->msg_id = ANKI_VEHICLE_MSG_C2V_SET_LIGHTS;
    msg->light_mask = mask;

    return sizeof(anki_vehicle_msg_set_lights_t);
}

uint8_t anki_vehicle_msg_lights_pattern(anki_vehicle_msg_t *message, uint8_t channel, uint8_t effect, uint8_t start, uint8_t end, uint16_t cycles_per_min)
{
    assert(message != NULL);

    anki_vehicle_msg_lights_pattern_t *msg = (anki_vehicle_msg_lights_pattern_t *)message;
    memset(msg, 0, sizeof(anki_vehicle_msg_lights_pattern_t));
    msg->size = ANKI_VEHICLE_MSG_C2V_LIGHTS_PATTERN_SIZE;
    msg->msg_id = ANKI_VEHICLE_MSG_C2V_LIGHTS_PATTERN;
    msg->channel = channel;
    msg->effect = effect;
    msg->start = start;
    msg->end = end;
    msg->cycles_per_min = cycles_per_min;

    return sizeof(anki_vehicle_msg_lights_pattern_t);
}

uint8_t anki_vehicle_msg_disconnect(anki_vehicle_msg_t *msg)
{
    assert(msg != NULL);
    msg->size = ANKI_VEHICLE_MSG_BASE_SIZE;
    msg->msg_id = ANKI_VEHICLE_MSG_C2V_DISCONNECT;
    return ANKI_VEHICLE_MSG_TYPE_SIZE;
}

uint8_t anki_vehicle_msg_cancel_lane_change(anki_vehicle_msg_t *msg)
{
    assert(msg != NULL);
    msg->size = ANKI_VEHICLE_MSG_BASE_SIZE;
    msg->msg_id = ANKI_VEHICLE_MSG_C2V_CANCEL_LANE_CHANGE;
    return ANKI_VEHICLE_MSG_TYPE_SIZE;
}

uint8_t anki_vehicle_msg_turn_180(anki_vehicle_msg_t *msg)
{
    assert(msg != NULL);
    msg->size = ANKI_VEHICLE_MSG_BASE_SIZE;
    msg->msg_id = ANKI_VEHICLE_MSG_C2V_TURN_180;
    return ANKI_VEHICLE_MSG_TYPE_SIZE;
}

uint8_t anki_vehicle_msg_ping(anki_vehicle_msg_t *msg)
{
    assert(msg != NULL);
    msg->size = ANKI_VEHICLE_MSG_BASE_SIZE;
    msg->msg_id = ANKI_VEHICLE_MSG_C2V_PING_REQUEST;
    return ANKI_VEHICLE_MSG_TYPE_SIZE;
}

uint8_t anki_vehicle_msg_get_version(anki_vehicle_msg_t *msg)
{
    assert(msg != NULL);
    msg->size = ANKI_VEHICLE_MSG_BASE_SIZE;
    msg->msg_id = ANKI_VEHICLE_MSG_C2V_VERSION_REQUEST;
    return ANKI_VEHICLE_MSG_TYPE_SIZE;
}

uint8_t anki_vehicle_msg_get_battery_level(anki_vehicle_msg_t *msg)
{
    assert(msg != NULL);
    msg->size = ANKI_VEHICLE_MSG_BASE_SIZE;
    msg->msg_id = ANKI_VEHICLE_MSG_C2V_BATTERY_LEVEL_REQUEST;
    return ANKI_VEHICLE_MSG_TYPE_SIZE;
}
