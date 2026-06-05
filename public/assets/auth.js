const titles = {
  login: {
    title: "Connexion",
    description: "Connecte-toi pour publier, sauvegarder et repondre aux questions."
  },
  register: {
    title: "Inscription",
    description: "Cree ton compte pour rejoindre la communaute et poser tes questions."
  },
  forgot: {
    title: "Mot de passe oublie",
    description: "Demande un lien de reinitialisation si tu n'as plus acces a ton compte."
  },
  verify: {
    title: "Verification du compte",
    description: "Active ton compte avec le code recu par email apres inscription."
  },
  reset: {
    title: "Nouveau mot de passe",
    description: "Definis un nouveau mot de passe a partir du token recu par email."
  }
};

const alertBox = document.querySelector("#auth-alert");
const panelTitle = document.querySelector("#panel-title");
const panelDescription = document.querySelector("#panel-description");
const tabs = document.querySelectorAll("[data-mode]");
const forms = document.querySelectorAll("[data-form]");
let captchaWidgetId = null;

function setMode(mode) {
  const copy = titles[mode] || titles.login;

  panelTitle.textContent = copy.title;
  panelDescription.textContent = copy.description;
  alertBox.hidden = true;

  forms.forEach((form) => {
    form.classList.toggle("active", form.dataset.form === mode);
  });

  document.querySelectorAll(".mode-tab").forEach((tab) => {
    const isActive = tab.dataset.mode === mode;
    tab.classList.toggle("active", isActive);
    tab.setAttribute("aria-selected", String(isActive));
  });

  const params = new URLSearchParams(window.location.search);
  params.set("mode", mode);
  window.history.replaceState({}, "", `${window.location.pathname}?${params.toString()}`);
}

function showAlert(message, type = "success") {
  alertBox.textContent = message;
  alertBox.className = `alert ${type}`;
  alertBox.hidden = false;
}

function formDataToJson(form) {
  return Object.fromEntries(new FormData(form).entries());
}

function isRecaptchaConfigured() {
  return window.RECAPTCHA_SITE_KEY && window.RECAPTCHA_SITE_KEY !== "YOUR_RECAPTCHA_SITE_KEY";
}

function syncCaptchaToken(token = "") {
  const captchaInput = document.querySelector("#login-form input[name='captcha']");
  captchaInput.value = token;
}

window.onRecaptchaLoaded = function onRecaptchaLoaded() {
  if (!isRecaptchaConfigured()) {
    showAlert("Configure RECAPTCHA_SITE_KEY dans assets/auth-config.js pour activer le captcha.", "error");
    return;
  }

  captchaWidgetId = grecaptcha.render("login-captcha", {
    sitekey: window.RECAPTCHA_SITE_KEY,
    callback: syncCaptchaToken,
    "expired-callback": () => syncCaptchaToken(""),
    "error-callback": () => syncCaptchaToken("")
  });
};

function resetCaptcha() {
  syncCaptchaToken("");

  if (captchaWidgetId !== null && window.grecaptcha) {
    grecaptcha.reset(captchaWidgetId);
  }
}

async function requestJson(url, options = {}) {
  const response = await fetch(url, {
    headers: { "Content-Type": "application/json" },
    ...options
  });

  const text = await response.text();
  let payload = {};

  try {
    payload = text ? JSON.parse(text) : {};
  } catch {
    payload = { message: text };
  }

  if (!response.ok) {
    throw new Error(payload.message || text || "Une erreur est survenue.");
  }

  return payload;
}

async function submitForm(form, handler) {
  const button = form.querySelector("button[type='submit']");
  button.disabled = true;

  try {
    await handler(form);
  } catch (error) {
    showAlert(error.message, "error");
  } finally {
    button.disabled = false;
  }
}

document.querySelector("#login-form").addEventListener("submit", (event) => {
  event.preventDefault();
  submitForm(event.currentTarget, async (form) => {
    if (!isRecaptchaConfigured()) {
      throw new Error("Le captcha n'est pas configure. Renseigne la cle publique reCAPTCHA.");
    }

    const data = formDataToJson(form);

    if (!data.captcha) {
      throw new Error("Valide le captcha avant de te connecter.");
    }

    const payload = await requestJson("/api/login", {
      method: "POST",
      body: JSON.stringify(data)
    });

    localStorage.setItem("access_token", payload.token);
    localStorage.setItem("refresh_token", payload.refresh_token);
    showAlert("Connexion reussie. Redirection vers l'accueil...");
    window.setTimeout(() => window.location.assign("index.html"), 700);
  });
  resetCaptcha();
});

document.querySelector("#register-form").addEventListener("submit", (event) => {
  event.preventDefault();
  submitForm(event.currentTarget, async (form) => {
    const payload = await requestJson("/api/register", {
      method: "POST",
      body: JSON.stringify(formDataToJson(form))
    });

    showAlert(payload.message || "Compte cree. Verifie ta boite mail pour l'activer.");
    setMode("verify");
    document.querySelector("#verify-form input[name='email']").value = form.email.value;
  });
});

document.querySelector("#verify-form").addEventListener("submit", (event) => {
  event.preventDefault();
  submitForm(event.currentTarget, async (form) => {
    const data = formDataToJson(form);
    const params = new URLSearchParams(data);
    const payload = await requestJson(`/api/verify-register?${params.toString()}`);

    showAlert(payload.message || "Compte verifie. Tu peux maintenant te connecter.");
    setMode("login");
    document.querySelector("#login-form input[name='email']").value = data.email;
  });
});

document.querySelector("#forgot-form").addEventListener("submit", (event) => {
  event.preventDefault();
  submitForm(event.currentTarget, async (form) => {
    const payload = await requestJson("/api/forgot-password", {
      method: "POST",
      body: JSON.stringify(formDataToJson(form))
    });

    showAlert(payload.message || "Si cet email existe, un lien a ete envoye.");
  });
});

document.querySelector("#reset-form").addEventListener("submit", (event) => {
  event.preventDefault();
  submitForm(event.currentTarget, async (form) => {
    const payload = await requestJson("/api/confirm-reset", {
      method: "POST",
      body: JSON.stringify(formDataToJson(form))
    });

    showAlert(payload.message || "Mot de passe mis a jour. Tu peux te connecter.");
    setMode("login");
  });
});

tabs.forEach((tab) => {
  tab.addEventListener("click", () => setMode(tab.dataset.mode));
});

const initialParams = new URLSearchParams(window.location.search);
const token = initialParams.get("token");

if (token) {
  setMode("reset");
  document.querySelector("#reset-form input[name='token']").value = token;
} else {
  setMode(initialParams.get("mode") || "login");
}
