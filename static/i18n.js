(function (global) {
  "use strict";

  const STORAGE_KEY = "pastebin-lang";

  const messages = {
    en: {
      "meta.title": "GFONSECA Pastebin — Smart Code Sharing",
      "meta.description": "Paste, share, and view code, JSON, and logs in seconds.",
      "meta.viewTitle": "Paste {id} — Viewer",
      "badge.online": "v1.0 · Online",
      "tagline":
        "The fastest, most elegant way to share logs, code, JSON, and scripts — with instant links.",
      "pill.instant": "Instant",
      "pill.secure": "Secure link",
      "pill.cli": "CLI · cURL & Wget",
      "pill.syntax": "Syntax highlight",
      "editor.status": "Server ready",
      "editor.placeholder":
        "// Paste or type your code here…\n// Ctrl+Enter to generate the link",
      "editor.chars": "{n} characters",
      "editor.line": "{n} line",
      "editor.lines": "{n} lines",
      "btn.create": "Create secure link",
      "btn.generating": "Generating link…",
      "cli.title": "Via terminal (POST)",
      "cli.example":
        "cat file | curl --data-binary @- https://paste.gfonseca.online/paste",
      "result.title": "Link generated successfully",
      "result.subtitle": "Your content has been saved and is ready to share.",
      "btn.copyLink": "Copy link",
      "btn.copied": "Copied!",
      "share.title": "Download commands (click to copy)",
      "share.copy": "copy",
      "share.copyTitle": "Click to copy",
      "share.raw": "View plain text (RAW)",
      "footer.madeWith": "Built with Go ·",
      "footer.createNew": "Create new paste",
      "view.title": "View",
      "view.titleAccent": " Paste",
      "btn.raw": "Plain text (RAW)",
      "btn.copyCode": "Copy code",
      "meta.bytesLines": "{bytes} bytes · {lines}",
      "meta.line": "line",
      "meta.lines": "lines",
      "toast.error": "Could not save the paste. Please try again.",
      "lang.label": "Language",
    },
    pt: {
      "meta.title": "GFONSECA Pastebin — Compartilhamento Inteligente de Código",
      "meta.description":
        "Cole, compartilhe e visualize código, JSON e logs em segundos.",
      "meta.viewTitle": "Paste {id} — Visualizador",
      "badge.online": "v1.0 · Online",
      "tagline":
        "A forma mais rápida e elegante de compartilhar logs, códigos, JSON e scripts — com link instantâneo.",
      "pill.instant": "Instantâneo",
      "pill.secure": "Link seguro",
      "pill.cli": "CLI · cURL & Wget",
      "pill.syntax": "Syntax highlight",
      "editor.status": "Servidor pronto",
      "editor.placeholder":
        "// Cole ou digite seu código aqui…\n// Ctrl+Enter para gerar o link",
      "editor.chars": "{n} caracteres",
      "editor.line": "{n} linha",
      "editor.lines": "{n} linhas",
      "btn.create": "Criar link seguro",
      "btn.generating": "Gerando link…",
      "cli.title": "Via terminal (POST)",
      "cli.example":
        "cat arquivo | curl --data-binary @- https://paste.gfonseca.online/paste",
      "result.title": "Link gerado com sucesso",
      "result.subtitle": "Seu conteúdo foi salvo e está pronto para compartilhar.",
      "btn.copyLink": "Copiar link",
      "btn.copied": "Copiado!",
      "share.title": "Comandos para download (clique para copiar)",
      "share.copy": "copiar",
      "share.copyTitle": "Clique para copiar",
      "share.raw": "Visualizar texto puro (RAW)",
      "footer.madeWith": "Feito com Go ·",
      "footer.createNew": "Criar novo paste",
      "view.title": "Visualizar",
      "view.titleAccent": " Paste",
      "btn.raw": "Texto puro (RAW)",
      "btn.copyCode": "Copiar código",
      "meta.bytesLines": "{bytes} bytes · {lines}",
      "meta.line": "linha",
      "meta.lines": "linhas",
      "toast.error": "Não foi possível salvar o paste. Tente novamente.",
      "lang.label": "Idioma",
    },
  };

  let currentLang = "en";
  const listeners = [];

  function detectLanguage() {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved === "en" || saved === "pt") return saved;

    const nav = (navigator.language || navigator.userLanguage || "en").toLowerCase();
    return nav.startsWith("pt") ? "pt" : "en";
  }

  function format(key, vars) {
    let text = messages[currentLang][key] || messages.en[key] || key;
    if (vars) {
      Object.entries(vars).forEach(([k, v]) => {
        text = text.replace(new RegExp(`\\{${k}\\}`, "g"), String(v));
      });
    }
    return text;
  }

  function t(key, vars) {
    return format(key, vars);
  }

  function localeTag() {
    return currentLang === "pt" ? "pt-BR" : "en";
  }

  function numberLocale() {
    return currentLang === "pt" ? "pt-BR" : "en-US";
  }

  function applyLanguage(lang) {
    if (!messages[lang]) lang = "en";
    currentLang = lang;
    localStorage.setItem(STORAGE_KEY, lang);

    document.documentElement.lang = localeTag();

    document.querySelectorAll("[data-i18n]").forEach((el) => {
      el.textContent = t(el.dataset.i18n);
    });

    document.querySelectorAll("[data-i18n-placeholder]").forEach((el) => {
      el.placeholder = t(el.dataset.i18nPlaceholder);
    });

    document.querySelectorAll("[data-i18n-title]").forEach((el) => {
      el.title = t(el.dataset.i18nTitle);
    });

    const metaDesc = document.querySelector('meta[name="description"]');
    if (metaDesc && document.body.dataset.page === "index") {
      metaDesc.content = t("meta.description");
    }

    if (document.body.dataset.page === "index") {
      document.title = t("meta.title");
    } else if (document.body.dataset.page === "view") {
      const id = document.body.dataset.pasteId || "";
      document.title = t("meta.viewTitle", { id });
    }

    document.querySelectorAll(".lang-btn").forEach((btn) => {
      const active = btn.dataset.lang === currentLang;
      btn.classList.toggle("lang-btn--active", active);
      btn.setAttribute("aria-pressed", active ? "true" : "false");
    });

    listeners.forEach((fn) => fn(currentLang));
  }

  function setLanguage(lang) {
    applyLanguage(lang);
  }

  function onChange(fn) {
    listeners.push(fn);
  }

  function initLanguageSwitcher() {
    document.querySelectorAll(".lang-btn").forEach((btn) => {
      btn.addEventListener("click", () => setLanguage(btn.dataset.lang));
    });
  }

  function init() {
    applyLanguage(detectLanguage());
    initLanguageSwitcher();
  }

  global.PastebinI18n = {
    t,
    format,
    applyLanguage,
    setLanguage,
    onChange,
    getLanguage: () => currentLang,
    numberLocale,
    init,
  };

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})(window);
