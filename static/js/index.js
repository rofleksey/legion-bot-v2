const { createApp, ref, onMounted, onUnmounted } = Vue;

createApp({
  setup() {
    // Reactive state
    const user = ref(null);
    const featuresVisible = ref(Array(6).fill(false));
    const isLoading = ref(false);

    // Methods
    const loginWithTwitch = async () => {
      isLoading.value = true;
      try {
        const response = await axios.get('/api/auth/login');
        if (response.data?.authUrl) {
          localStorage.setItem('twitch_auth_state', response.data.state);
          window.location.href = response.data.authUrl;
        }
      } catch (error) {
        console.error('Error during Twitch login:', error);
        // Consider adding user feedback here
      } finally {
        isLoading.value = false;
      }
    };

    const validateToken = async () => {
      try {
        const token = localStorage.getItem('legionbot_token');
        if (!token) return;

        const response = await axios.get('/api/validate', {
          headers: { 'Authorization': `Bearer ${token}` }
        });

        if (response.data) {
          user.value = response.data;
        }
      } catch (error) {
        console.error('Token validation failed:', error);
        // Clear invalid token
        localStorage.removeItem('legionbot_token');
      }
    };

    const handleScroll = () => {
      const featureElements = document.querySelectorAll('.feature-card');
      featureElements.forEach((el, index) => {
        if (el) {
          const rect = el.getBoundingClientRect();
          featuresVisible.value[index] = rect.top < window.innerHeight - 100;
        }
      });
    };

    // Lifecycle hooks
    onMounted(() => {
      // Handle OAuth callback
      const urlParams = new URLSearchParams(window.location.search);
      const token = urlParams.get('token');
      const state = urlParams.get('state');

      if (token && state) {
        const storedState = localStorage.getItem('twitch_auth_state');
        if (state === storedState) {
          localStorage.setItem('legionbot_token', token);
          localStorage.removeItem('twitch_auth_state');
          window.history.replaceState({}, '', window.location.pathname);
        } else {
          console.error('State mismatch - possible CSRF attack');
        }
      }

      // Always validate token (new or existing)
      validateToken();

      // Scroll handling
      window.addEventListener('scroll', handleScroll, { passive: true });
      handleScroll(); // Initial check
    });

    // Cleanup
    onUnmounted(() => {
      window.removeEventListener('scroll', handleScroll);
    });

    return {
      user,
      isLoading,
      featuresVisible,
      loginWithTwitch
    };
  }
}).mount('#app');
