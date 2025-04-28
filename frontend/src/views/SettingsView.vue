<template>
  <div class="settings-container">
    <div class="settings-header">
      <h1 class="settings-title">Settings</h1>
      <div class="settings-actions">
        <button
          class="settings-save-button"
          :disabled="isSaving"
          @click="saveSettings"
        >
          {{ isSaving ? 'Saving...' : 'Save Changes' }}
        </button>
      </div>
    </div>

    <div class="settings-content" v-if="settings">
      <div class="settings-section">
        <h2 class="settings-section-title">General Settings</h2>
        <div class="settings-grid">
          <AppSwitch
            :model-value="!settings.disabled"
            @update:model-value="updateDisabled"
            :label="settings.disabled ? 'Disabled' : 'Enabled'"
          />

          <AppSelect
            v-model="settings.language"
            label="Language"
            :options="['en', 'ru']"
          />
        </div>
      </div>

      <div class="settings-section">
        <h2 class="settings-section-title">Killers Settings</h2>
        <div class="settings-subsection">
          <h3 class="settings-subsection-title">General</h3>
          <div class="settings-grid">
            <AppDurationInput
              v-model="settings.killers.general.delayBetweenKillers"
              :min="1e9"
              label="Delay Between Killers"
            />
          </div>
        </div>

        <div class="settings-subsection">
          <h3 class="settings-subsection-title">Legion</h3>
          <div class="settings-grid">
            <AppSwitch
              v-model="settings.killers.legion.enabled"
              :label="settings.killers.legion.enabled ? 'Enabled' : 'Disabled'"
            />
          </div>
          <div class="settings-grid" :class="{disabled: !settings.killers.legion.enabled}">
            <AppNumberInput
              v-model="settings.killers.legion.fatalHit"
              :min="2"
              label="Fatal Hit"
            />
            <AppDurationInput
              v-model="settings.killers.legion.frenzyTimeout"
              label="Frenzy Timeout"
            />
            <AppDurationInput
              v-model="settings.killers.legion.deepWoundTimeout"
              label="Deep Wound Timeout"
            />
            <AppChanceInput
              v-model="settings.killers.legion.reactChance"
              label="Message React Chance"
            />
            <AppChanceInput
              v-model="settings.killers.legion.hitChance"
              label="Hit Chance"
            />
            <AppDurationInput
              v-model="settings.killers.legion.minDelayBetweenHits"
              label="Min Delay Between Hits"
            />
            <AppDurationInput
              v-model="settings.killers.legion.hookBanTime"
              label="Hook Ban Time"
            />
            <AppDurationInput
              v-model="settings.killers.legion.bleedOutBanTime"
              label="Bleed Out Ban Time"
            />
            <AppChanceInput
              v-model="settings.killers.legion.bodyBlockSuccessChance"
              label="Body Block Success Chance"
            />
            <AppChanceInput
              v-model="settings.killers.legion.lockerGrabChance"
              label="Locker Grab Chance"
            />
            <AppChanceInput
              v-model="settings.killers.legion.lockerStunChance"
              label="Locker Stun Chance"
            />
            <AppChanceInput
              v-model="settings.killers.legion.palletStunChance"
              label="Pallet Stun Chance"
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

const notifications = useNotifications()

const userStore = useUserStore()
const token = computed(() => userStore.token)

const settings = ref<Settings | null>(null);

const isSaving = ref(false);

function updateDisabled(val: boolean) {
  if (!settings.value) {
    return
  }

  settings.value.disabled = !val
}

async function fetchSettings() {
  try {
    const response = await axios.get('/api/settings', {
      headers: {
        Authorization: `Bearer ${token.value}`
      }
    });
    settings.value = response.data;
  } catch (error) {
    console.error('Failed to fetch settings:', error);
  }
}

async function saveSettings() {
  isSaving.value = true;
  try {
    await axios.post('/api/settings', settings.value, {
      headers: {
        Authorization: `Bearer ${token.value}`
      }
    });
    notifications.info('Settings applied successfully!', 'OK')
  } catch (error) {
    notifications.error('Failed to apply settings', error?.toString?.() ?? '');
  } finally {
    isSaving.value = false;
  }
}

onMounted(() => {
  fetchSettings();
});
</script>

<style scoped>
.settings-container {
  max-width: 1200px;
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

.settings-save-button {
  padding: 8px 16px;
  background-color: var(--primary);
  color: var(--primary-foreground);
  border: none;
  border-radius: 4px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s;
}

.settings-save-button:hover {
  background-color: var(--primary-dark);
}

.settings-save-button:disabled {
  background-color: var(--muted);
  cursor: not-allowed;
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

.settings-grid {
  transition: opacity 0.3s ease;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}

.settings-grid.disabled {
  opacity: 0.4;
  pointer-events: none;
}
</style>
