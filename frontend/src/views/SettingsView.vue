<template>
  <div class="settings-container">
    <div class="settings-header">
      <h1 class="settings-title">{{ t('settings.title') }}</h1>
      <div class="settings-actions">
        <AppButton :loading="postLoading"  @click="saveSettings">
          {{ postLoading ? t('settings.saving') : t('settings.save_button') }}
        </AppButton>
      </div>
    </div>

    <div class="settings-content" v-if="settings">
      <div class="settings-section">
        <h2 class="settings-section-title">{{ t('settings.general_title') }}</h2>
        <div class="settings-grid">
          <AppSwitch
            :model-value="!settings.disabled"
            @update:model-value="updateDisabled"
            :label="settings.disabled ? t('settings.disabled') : t('settings.enabled')"
          />

          <AppSelect
            v-model="settings.language"
            :label="t('settings.language')"
            :options="['en', 'ru']"
          />
        </div>
      </div>

      <div class="settings-section">
        <h2 class="settings-section-title">{{ t('settings.killers_title') }}</h2>
        <div class="settings-subsection">
          <h3 class="settings-subsection-title">{{ t('settings.general') }}</h3>
          <div class="settings-grid">
            <AppDurationInput
              v-model="settings.killers.general.delayBetweenKillers"
              :min="1e9"
              :label="t('settings.delay_between_killers')"
            />
            <AppDurationInput
              v-model="settings.killers.general.delayAtTheStreamStart"
              :min="300 * 1e9"
              :label="t('settings.delay_at_the_stream_start')"
            />
            <AppNumberInput
              v-model="settings.killers.general.minNumberOfViewers"
              :min="0"
              :label="t('settings.min_number_of_viewers')"
            />
          </div>
        </div>

        <div class="settings-subsection">
          <h3 class="settings-subsection-title">{{ t('settings.legion') }}</h3>
          <AppQuotation class="settings-subsection-description">{{ t('settings.legion_description') }}</AppQuotation>
          <div class="settings-grid">
            <AppSwitch
              v-model="settings.killers.legion.enabled"
              :label="settings.killers.legion.enabled ? t('settings.enabled') : t('settings.disabled')"
            />
            <AppNumberInput
              v-model="settings.killers.legion.weight"
              :min="1"
              :max="1000000"
              :label="t('settings.weight')"
              show-help-icon
              @help-click="Dialog.show(t('settings.weight_description'))"
            />
            <AppButton :loading="postLoading" @click="summonKiller('legion')">
              {{ t('settings.summon') }}
            </AppButton>
          </div>
          <div class="settings-grid">
            <AppNumberInput
              v-model="settings.killers.legion.fatalHit"
              :min="2"
              :label="t('settings.fatal_hit')"
            />
            <AppDurationInput
              v-model="settings.killers.legion.frenzyTimeout"
              :label="t('settings.frenzy_timeout')"
            />
            <AppDurationInput
              v-model="settings.killers.legion.deepWoundTimeout"
              :label="t('settings.deep_wound_timeout')"
            />
            <AppChanceInput
              v-model="settings.killers.legion.reactChance"
              :label="t('settings.react_chance')"
            />
            <AppDurationInput
              v-model="settings.killers.legion.minDelayBetweenHits"
              :label="t('settings.min_delay_between_hits')"
            />
            <AppChanceInput
              v-model="settings.killers.legion.hitChance"
              :label="t('settings.hit_chance')"
            />
            <AppDurationInput
              v-model="settings.killers.legion.hookBanTime"
              :label="t('settings.hook_ban_time')"
            />
            <AppDurationInput
              v-model="settings.killers.legion.bleedOutBanTime"
              :label="t('settings.bleedout_ban_time')"
            />
            <AppChanceInput
              v-model="settings.killers.legion.bodyBlockSuccessChance"
              :label="t('settings.bodyblock_chance')"
            />
            <AppChanceInput
              v-model="settings.killers.legion.lockerGrabChance"
              :label="t('settings.locker_grab_chance')"
            />
            <AppChanceInput
              v-model="settings.killers.legion.lockerStunChance"
              :label="t('settings.locker_stun_chance')"
            />
            <AppChanceInput
              v-model="settings.killers.legion.palletStunChance"
              :label="t('settings.pallet_stun_chance')"
            />
          </div>
        </div>

        <div class="settings-subsection">
          <h3 class="settings-subsection-title">{{ t('settings.ghostface') }}</h3>
          <AppQuotation class="settings-subsection-description">{{ t('settings.ghostface_description') }}</AppQuotation>
          <div class="settings-grid">
            <AppSwitch
              v-model="settings.killers.ghostface.enabled"
              :label="settings.killers.ghostface.enabled ? t('settings.enabled') : t('settings.disabled')"
            />
            <AppNumberInput
              v-model="settings.killers.ghostface.weight"
              :min="1"
              :max="1000000"
              :label="t('settings.weight')"
              show-help-icon
              @help-click="Dialog.show(t('settings.weight_description'))"
            />
            <AppButton :loading="postLoading" @click="summonKiller('ghostface')">
              {{ t('settings.summon') }}
            </AppButton>
          </div>
          <div class="settings-grid">
            <AppDurationInput
              v-model="settings.killers.ghostface.timeout"
              :label="t('settings.timeout')"
            />
            <AppChanceInput
              v-model="settings.killers.ghostface.reactChance"
              :label="t('settings.react_chance')"
            />
            <AppDurationInput
              v-model="settings.killers.ghostface.minDelayBetweenHits"
              :label="t('settings.min_delay_between_hits')"
            />
            <AppDurationInput
              v-model="settings.killers.ghostface.hookBanTime"
              :label="t('settings.hook_ban_time')"
            />
          </div>
        </div>

        <div class="settings-subsection">
          <h3 class="settings-subsection-title">{{ t('settings.doctor') }}</h3>
          <AppQuotation class="settings-subsection-description">{{ t('settings.doctor_description') }}</AppQuotation>
          <div class="settings-grid">
            <AppSwitch
              v-model="settings.killers.doctor.enabled"
              :label="settings.killers.doctor.enabled ? t('settings.enabled') : t('settings.disabled')"
            />
            <AppNumberInput
              v-model="settings.killers.doctor.weight"
              :min="1"
              :max="1000000"
              :label="t('settings.weight')"
              show-help-icon
              @help-click="Dialog.show(t('settings.weight_description'))"
            />
            <AppButton :loading="postLoading" @click="summonKiller('doctor')">
              {{ t('settings.summon') }}
            </AppButton>
          </div>
          <div class="settings-grid">
            <AppDurationInput
              v-model="settings.killers.doctor.timeout"
              :label="t('settings.timeout')"
            />
            <AppChanceInput
              v-model="settings.killers.doctor.reactChance"
              :label="t('settings.react_chance')"
            />
            <AppDurationInput
              v-model="settings.killers.doctor.minDelayBetweenHits"
              :label="t('settings.min_delay_between_hits')"
            />
          </div>
        </div>

        <div class="settings-subsection">
          <h3 class="settings-subsection-title">{{ t('settings.pinhead') }}</h3>
          <AppQuotation class="settings-subsection-description">{{ t('settings.pinhead_description') }}</AppQuotation>
          <div class="settings-grid">
            <AppSwitch
              v-model="settings.killers.pinhead.enabled"
              :label="settings.killers.pinhead.enabled ? t('settings.enabled') : t('settings.disabled')"
            />
            <AppNumberInput
              v-model="settings.killers.pinhead.weight"
              :min="1"
              :max="1000000"
              :label="t('settings.weight')"
              show-help-icon
              @help-click="Dialog.show(t('settings.weight_description'))"
            />
            <AppButton :loading="postLoading" @click="summonKiller('pinhead')">
              {{ t('settings.summon') }}
            </AppButton>
          </div>
          <div class="settings-grid">
            <AppSwitch
              v-model="settings.killers.pinhead.showTopic"
              :label="settings.killers.pinhead.showTopic ? t('settings.show_topic') : t('settings.hide_topic')"
            />
            <AppDurationInput
              v-model="settings.killers.pinhead.deepWoundTimeout"
              :label="t('settings.deep_wound_timeout')"
            />
            <AppDurationInput
              v-model="settings.killers.pinhead.bleedOutBanTime"
              :label="t('settings.bleedout_ban_time')"
            />
            <AppDurationInput
              v-model="settings.killers.pinhead.timeout"
              :label="t('settings.timeout')"
            />
            <AppNumberInput
              v-model="settings.killers.pinhead.victimCount"
              :label="t('settings.victim_count')"
            />
            <AppStringInput
              v-model="settings.killers.pinhead.topics"
              :label="t('settings.topics')"
            />
          </div>
        </div>
      </div>

      <div class="settings-section">
        <h2 class="settings-section-title">{{ t('settings.chat_title') }}</h2>
        <div class="settings-subsection">
          <h3 class="settings-subsection-title">{{ t('settings.raids') }}</h3>
          <div class="settings-grid">
            <AppSwitch
              v-model="settings.chat.startKillerOnRaid"
              :label="t('settings.start_killer_on_raids')"
              show-help-icon
              @help-click="Dialog.show(t('settings.start_killer_on_raids_info'))"
            />
            <AppSwitch
              v-model="settings.chat.followRaids"
              :label="t('settings.follow_raids')"
            />
            <AppStringInput
              v-model="settings.chat.followRaidsMessage"
              :label="t('settings.follow_raids_message')"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {ref, onMounted, computed} from 'vue';
import axios from 'axios';
import {useUserStore} from "@/stores/user";
import type {Settings} from "@/lib/types";
import AppSwitch from "@/components/AppSwitch.vue";
import AppSelect from "@/components/AppSelect.vue";
import AppDurationInput from "@/components/AppDurationInput.vue";
import AppNumberInput from "@/components/AppNumberInput.vue";
import AppChanceInput from "@/components/AppChanceInput.vue";
import {useNotifications} from "@/services/notifications.ts";
import {useI18n} from "vue-i18n";
import AppQuotation from "@/components/AppQuotation.vue";
import AppStringInput from "@/components/AppStringInput.vue";
import {Dialog} from "@/services/dialog.ts";
import AppButton from "@/components/AppButton.vue";
import {errorToString} from "@/lib/misc.ts";

const {t} = useI18n()
const notifications = useNotifications()

const userStore = useUserStore()
const token = computed(() => userStore.token)

const settings = ref<Settings | null>(null);
const postLoading = ref(false);

function updateDisabled(val: boolean) {
  if (!settings.value) return;
  settings.value.disabled = !val;
}

async function fetchSettings() {
  try {
    const response = await axios.get('/api/settings', {
      headers: {Authorization: `Bearer ${token.value}`}
    });
    settings.value = response.data;
  } catch (error) {
    console.error('Failed to fetch settings:', error);
  }
}

async function saveSettings() {
  postLoading.value = true;
  try {
    await axios.post('/api/settings', settings.value, {
      headers: {Authorization: `Bearer ${token.value}`}
    });
    notifications.info(t('settings.save_success'), 'OK')
  } catch (e) {
    notifications.error(t('settings.save_failed'), errorToString(e));
  } finally {
    postLoading.value = false;
  }
}

async function summonKiller(name: string) {
  postLoading.value = true;
  try {
    await axios.post('/api/summonKiller', { name }, {
      headers: {Authorization: `Bearer ${token.value}`}
    });
    notifications.info(t('settings.summoned'), 'OK')
  } catch (e) {
    notifications.error(t('settings.summon_failed'), errorToString(e));
  } finally {
    postLoading.value = false;
  }
}

onMounted(fetchSettings);
</script>

<style scoped>
.settings-container {
  max-width: 1000px;
  margin: 0 auto;
  padding: 24px;
}

.settings-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 32px;
}

.settings-title {
  font-size: 24px;
  font-weight: 600;
  color: var(--foreground);
}

.settings-actions {
  display: flex;
  gap: 12px;
}

.settings-content {
  display: flex;
  flex-direction: column;
  gap: 32px;
}

.settings-section {
  background-color: var(--card);
  border-radius: 8px;
  padding: 24px;
  box-shadow: var(--shadow-sm);
}

.settings-section-title {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 20px;
  color: var(--foreground);
}

.settings-subsection {
  margin-top: 24px;
}

.settings-subsection-title {
  font-size: 18px;
  font-weight: 500;
  margin-bottom: 16px;
  color: var(--foreground);
}

.settings-subsection-description {
  font-size: 15px;
  font-weight: 500;
  margin-bottom: 16px;
  color: var(--foreground);
}

.settings-grid {
  transition: opacity 0.3s ease;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(275px, 1fr));
  gap: 16px;
  margin-bottom: 32px;
  justify-content: start;
  justify-items: start;
  align-items: center;
  align-content: start;
}

.settings-grid.disabled {
  opacity: 0.4;
  pointer-events: none;
}
</style>
