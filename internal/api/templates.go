package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// handleUserPage 用户页面
func (s *Server) handleUserPage(c *gin.Context) {
	html := s.renderUserPage()
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// handleAdminPage 管理员页面
func (s *Server) handleAdminPage(c *gin.Context) {
	html := s.renderAdminPage()
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// renderUserPage 渲染用户页面
func (s *Server) renderUserPage() string {
	// 构建 OAuth URL
	authURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&scope=user",
		s.config.LinuxDoAuthURL,
		s.config.LinuxDoClientID,
		s.config.LinuxDoRedirectURI,
	)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>KYX API Quota Bridge</title>
  <style>
%s
  </style>
</head>
<body>
%s

  <script>
    let currentUser = null;

    // Toast 通知函数
    function showToast(title, message, type = 'success') {
      const toast = document.createElement('div');
      toast.className = 'toast ' + type;
      toast.innerHTML = `+"`"+`
        <div class="toast-icon">${type === 'success' ? '✅' : '❌'}</div>
        <div class="toast-content">
          <div class="toast-title">${title}</div>
          ${message ? '<div class="toast-message">' + message + '</div>' : ''}
        </div>
      `+"`"+`;
      document.body.appendChild(toast);
      
      setTimeout(() => {
        toast.style.animation = 'slideInRight 0.3s ease reverse';
        setTimeout(() => toast.remove(), 300);
      }, 3000);
    }

    function toggleDropdown() {
      document.getElementById('dropdownMenu').classList.toggle('show');
    }

    document.addEventListener('click', (e) => {
      const userMenu = document.querySelector('.user-menu');
      if (userMenu && !userMenu.contains(e.target)) {
        document.getElementById('dropdownMenu').classList.remove('show');
      }
    });

    function showMessage(text, type) {
      const msg = document.getElementById('message');
      msg.textContent = text;
      msg.className = 'message ' + type;
      msg.classList.remove('hidden');
      setTimeout(() => msg.classList.add('hidden'), 5000);
    }

    function login() {
      const btn = event.target.closest('button');
      btn.disabled = true;
      btn.innerHTML = '<div class="spinner"></div> 登录中...';
      sessionStorage.setItem('oauth_loading', 'true');
      window.location.href = '%s';
    }

    function showBindMessage(text, type) {
      const msg = document.getElementById('bindMessage');
      msg.textContent = text;
      msg.className = 'message ' + type;
      msg.classList.remove('hidden');
      setTimeout(() => msg.classList.add('hidden'), 5000);
    }

    async function bindAccount() {
      const username = document.getElementById('username').value.trim();
      if (!username) {
        showBindMessage('请输入用户名', 'error');
        return;
      }

      try {
        const res = await fetch('/api/auth/bind', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ username }),
        });
        const data = await res.json();
        
        if (data.success) {
          document.getElementById('bindSection').classList.add('hidden');
          document.getElementById('mainSection').classList.remove('hidden');
          document.getElementById('navbar').style.display = 'block';
          await loadQuota();
          showToast('绑定成功', '已成功绑定公益站账号', 'success');
        } else {
          showBindMessage(data.message, 'error');
        }
      } catch (e) {
        showBindMessage('绑定失败: ' + e.message, 'error');
      }
    }

%s
  </script>
</body>
</html>`, getUserPageCSS(), getUserPageHTML(), authURL, getUserPageJS())
}

// renderAdminPage 渲染管理员页面
func (s *Server) renderAdminPage() string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>管理员后台 - KYX API Quota Bridge</title>
  <style>
%s
  </style>
</head>
<body>
%s

  <script>
%s
  </script>
</body>
</html>`, getAdminPageCSS(), getAdminPageHTML(), getAdminPageJS())
}
