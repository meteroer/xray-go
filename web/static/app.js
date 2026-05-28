const I18N = {
  zh: {
    overview: '概览',
    subscriptions: '订阅',
    nodes: '节点',
    settings: '设置',
    login: '登录',
    register: '创建用户',
    username: '用户名',
    password: '密码',
    confirmPassword: '确认密码',
    submit: '提交',
    logout: '退出',
    proxyStatus: '代理状态',
    running: '运行中',
    stopped: '已停止',
    currentNode: '当前节点',
    httpPort: 'HTTP 端口',
    socksPort: 'SOCKS5 端口',
    routeMode: '路由模式',
    startProxy: '启动代理',
    stopProxy: '停止代理',
    testLatency: '测速',
    add: '添加',
    delete: '删除',
    refresh: '刷新',
    name: '名称',
    url: '地址',
    nodesCount: '节点数',
    lastUpdated: '最后更新',
    actions: '操作',
    addSubscription: '添加订阅',
    addNode: '添加节点',
    nodeLink: '节点链接',
    region: '地区',
    allRegions: '全部地区',
    latency: '延迟',
    select: '选择',
    language: '语言',
    global: '全局',
    whitelist: '白名单',
    blacklist: '黑名单',
    noData: '暂无数据',
    error: '错误',
    success: '成功',
    passwordMismatch: '两次密码不一致',
  },
  en: {
    overview: 'Overview',
    subscriptions: 'Subscriptions',
    nodes: 'Nodes',
    settings: 'Settings',
    login: 'Login',
    register: 'Register',
    username: 'Username',
    password: 'Password',
    confirmPassword: 'Confirm Password',
    submit: 'Submit',
    logout: 'Logout',
    proxyStatus: 'Proxy Status',
    running: 'Running',
    stopped: 'Stopped',
    currentNode: 'Current Node',
    httpPort: 'HTTP Port',
    socksPort: 'SOCKS5 Port',
    routeMode: 'Route Mode',
    startProxy: 'Start Proxy',
    stopProxy: 'Stop Proxy',
    testLatency: 'Test Latency',
    add: 'Add',
    delete: 'Delete',
    refresh: 'Refresh',
    name: 'Name',
    url: 'URL',
    nodesCount: 'Nodes',
    lastUpdated: 'Last Updated',
    actions: 'Actions',
    addSubscription: 'Add Subscription',
    addNode: 'Add Node',
    nodeLink: 'Node Link',
    region: 'Region',
    allRegions: 'All Regions',
    latency: 'Latency',
    select: 'Select',
    language: 'Language',
    global: 'Global',
    whitelist: 'Whitelist',
    blacklist: 'Blacklist',
    noData: 'No data',
    error: 'Error',
    success: 'Success',
    passwordMismatch: 'Passwords do not match',
  }
};

class App {
  constructor() {
    this.lang = localStorage.getItem('lang') || 'zh';
    this.token = localStorage.getItem('token');
    this.config = null;
    this.proxyStatus = null;
    this.currentPage = 'overview';
    this.init();
  }

  t(key) {
    return I18N[this.lang][key] || key;
  }

  async init() {
    // Check auth status
    const statusRes = await fetch('/api/auth/status');
    const status = await statusRes.json();
    
    if (!status.initialized) {
      this.showAuth('register');
    } else if (!this.token) {
      this.showAuth('login');
    } else {
      this.showApp();
    }
  }

  showAuth(mode) {
    document.getElementById('auth-page').classList.remove('hidden');
    document.getElementById('app-page').classList.add('hidden');
    
    const title = mode === 'register' ? this.t('register') : this.t('login');
    document.getElementById('auth-title').textContent = title;
    document.getElementById('auth-submit').textContent = this.t('submit');
    document.getElementById('confirm-password-group').classList.toggle('hidden', mode !== 'register');
    
    document.getElementById('auth-form').onsubmit = async (e) => {
      e.preventDefault();
      const username = document.getElementById('username').value;
      const password = document.getElementById('password').value;
      
      if (mode === 'register') {
        const confirm = document.getElementById('confirm-password').value;
        if (password !== confirm) {
          alert(this.t('passwordMismatch'));
          return;
        }
      }
      
      const endpoint = mode === 'register' ? '/api/auth/init' : '/api/auth/login';
      const res = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
      });
      
      const data = await res.json();
      if (res.ok) {
        this.token = data.token;
        localStorage.setItem('token', this.token);
        this.showApp();
      } else {
        alert(data.error || this.t('error'));
      }
    };
  }

  async showApp() {
    document.getElementById('auth-page').classList.add('hidden');
    document.getElementById('app-page').classList.remove('hidden');
    
    await this.loadConfig();
    this.renderSidebar();
    this.renderPage();
    this.startStatusPolling();
  }

  async loadConfig() {
    const res = await fetch('/api/config', {
      headers: { 'Authorization': `Bearer ${this.token}` }
    });
    this.config = await res.json();
  }

  async loadProxyStatus() {
    const res = await fetch('/api/proxy/status', {
      headers: { 'Authorization': `Bearer ${this.token}` }
    });
    this.proxyStatus = await res.json();
  }

  startStatusPolling() {
    setInterval(() => this.loadProxyStatus().then(() => this.updateStatusUI()), 3000);
  }

  updateStatusUI() {
    if (!this.proxyStatus) return;
    const statusEl = document.getElementById('proxy-status-badge');
    if (statusEl) {
      statusEl.textContent = this.proxyStatus.running ? this.t('running') : this.t('stopped');
      statusEl.className = `status-badge ${this.proxyStatus.running ? 'running' : 'stopped'}`;
    }
    const nodeEl = document.getElementById('current-node');
    if (nodeEl) nodeEl.textContent = this.proxyStatus.node || '-';
  }

  renderSidebar() {
    const items = [
      { key: 'overview', icon: '📊' },
      { key: 'subscriptions', icon: '📋' },
      { key: 'nodes', icon: '🔌' },
      { key: 'settings', icon: '⚙️' }
    ];
    
    const nav = document.getElementById('sidebar-nav');
    nav.innerHTML = items.map(item => `
      <div class="nav-item ${this.currentPage === item.key ? 'active' : ''}" onclick="app.navigate('${item.key}')">
        <span>${item.icon}</span>
        <span>${this.t(item.key)}</span>
      </div>
    `).join('');
  }

  navigate(page) {
    this.currentPage = page;
    this.renderSidebar();
    this.renderPage();
  }

  renderPage() {
    const content = document.getElementById('main-content');
    switch (this.currentPage) {
      case 'overview':
        content.innerHTML = this.renderOverview();
        break;
      case 'subscriptions':
        content.innerHTML = this.renderSubscriptions();
        break;
      case 'nodes':
        content.innerHTML = this.renderNodes();
        break;
      case 'settings':
        content.innerHTML = this.renderSettings();
        break;
    }
  }

  renderOverview() {
    return `
      <div class="top-bar">
        <h2>${this.t('overview')}</h2>
        <button class="lang-switch" onclick="app.toggleLang()">${this.lang === 'zh' ? 'EN' : '中'}</button>
      </div>
      <div class="card">
        <h3>${this.t('proxyStatus')}</h3>
        <div style="display:flex;align-items:center;gap:20px;margin-bottom:20px;">
          <span id="proxy-status-badge" class="status-badge ${this.proxyStatus?.running ? 'running' : 'stopped'}">
            ${this.proxyStatus?.running ? this.t('running') : this.t('stopped')}
          </span>
        </div>
        <div style="display:grid;grid-template-columns:1fr 1fr;gap:16px;margin-bottom:20px;">
          <div>
            <div style="font-size:12px;color:#666;margin-bottom:4px;">${this.t('currentNode')}</div>
            <div id="current-node" style="font-weight:600;">${this.proxyStatus?.node || '-'}</div>
          </div>
          <div>
            <div style="font-size:12px;color:#666;margin-bottom:4px;">${this.t('routeMode')}</div>
            <div style="font-weight:600;">${this.proxyStatus?.route_mode || '-'}</div>
          </div>
          <div>
            <div style="font-size:12px;color:#666;margin-bottom:4px;">${this.t('httpPort')}</div>
            <div style="font-weight:600;">${this.proxyStatus?.http_port || '-'}</div>
          </div>
          <div>
            <div style="font-size:12px;color:#666;margin-bottom:4px;">${this.t('socksPort')}</div>
            <div style="font-weight:600;">${this.proxyStatus?.socks_port || '-'}</div>
          </div>
        </div>
        <div style="display:flex;gap:10px;">
          <button class="btn btn-success" onclick="app.startProxy()">${this.t('startProxy')}</button>
          <button class="btn btn-danger" onclick="app.stopProxy()">${this.t('stopProxy')}</button>
        </div>
      </div>
    `;
  }

  renderSubscriptions() {
    const subs = this.config?.subscriptions || [];
    return `
      <div class="top-bar">
        <h2>${this.t('subscriptions')}</h2>
        <button class="btn btn-primary" onclick="app.showAddSubModal()">+ ${this.t('addSubscription')}</button>
      </div>
      <div class="card">
        <table>
          <thead>
            <tr>
              <th>${this.t('name')}</th>
              <th>${this.t('url')}</th>
              <th>${this.t('nodesCount')}</th>
              <th>${this.t('lastUpdated')}</th>
              <th>${this.t('actions')}</th>
            </tr>
          </thead>
          <tbody>
            ${subs.length === 0 ? `<tr><td colspan="5" style="text-align:center;color:#999;">${this.t('noData')}</td></tr>` : ''}
            ${subs.map(sub => `
              <tr>
                <td>${sub.name}</td>
                <td style="max-width:300px;overflow:hidden;text-overflow:ellipsis;">${sub.url}</td>
                <td>${(sub.nodes || []).length}</td>
                <td>${sub.last_fetched ? new Date(sub.last_fetched).toLocaleString() : '-'}</td>
                <td>
                  <button class="btn btn-secondary" style="padding:4px 8px;font-size:12px;" onclick="app.refreshSub('${sub.name}')">${this.t('refresh')}</button>
                  <button class="btn btn-danger" style="padding:4px 8px;font-size:12px;" onclick="app.deleteSub('${sub.name}')">${this.t('delete')}</button>
                </td>
              </tr>
            `).join('')}
          </tbody>
        </table>
      </div>
    `;
  }

  renderNodes() {
    const standalone = this.config?.standalone_nodes || [];
    const subs = this.config?.subscriptions || [];
    let allNodes = [];
    subs.forEach(sub => {
      if (sub.nodes) allNodes = allNodes.concat(sub.nodes);
    });
    allNodes = allNodes.concat(standalone);

    return `
      <div class="top-bar">
        <h2>${this.t('nodes')}</h2>
        <div style="display:flex;gap:10px;">
          <button class="btn btn-secondary" onclick="app.testAllLatency()">${this.t('testLatency')}</button>
          <button class="btn btn-primary" onclick="app.showAddNodeModal()">+ ${this.t('addNode')}</button>
        </div>
      </div>
      <div class="card">
        <div id="nodes-list">
          ${allNodes.length === 0 ? `<div style="text-align:center;color:#999;padding:40px;">${this.t('noData')}</div>` : ''}
          ${allNodes.map(node => `
            <div class="node-item">
              <div>
                <div style="font-weight:600;">${node.name}</div>
                <div style="font-size:12px;color:#666;">${node.address}:${node.port} [${node.protocol}]</div>
              </div>
              <div style="display:flex;align-items:center;gap:10px;">
                <span class="latency-good" id="latency-${node.name}"></span>
                <button class="btn btn-success" style="padding:4px 12px;font-size:12px;" onclick="app.selectNode('${node.name}')">${this.t('select')}</button>
              </div>
            </div>
          `).join('')}
        </div>
      </div>
    `;
  }

  renderSettings() {
    return `
      <div class="top-bar">
        <h2>${this.t('settings')}</h2>
        <button class="lang-switch" onclick="app.toggleLang()">${this.lang === 'zh' ? 'EN' : '中'}</button>
      </div>
      <div class="card">
        <h3>${this.t('routeMode')}</h3>
        <div class="form-group">
          <select id="route-mode" onchange="app.changeRouteMode(this.value)">
            <option value="global" ${this.config?.route_mode === 'global' ? 'selected' : ''}>${this.t('global')}</option>
            <option value="whitelist" ${this.config?.route_mode === 'whitelist' ? 'selected' : ''}>${this.t('whitelist')}</option>
            <option value="blacklist" ${this.config?.route_mode === 'blacklist' ? 'selected' : ''}>${this.t('blacklist')}</option>
          </select>
        </div>
      </div>
      <div class="card">
        <button class="btn btn-secondary" onclick="app.logout()">${this.t('logout')}</button>
      </div>
    `;
  }

  async startProxy() {
    const res = await fetch('/api/proxy/start', {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${this.token}`, 'Content-Type': 'application/json' }
    });
    const data = await res.json();
    if (res.ok) {
      await this.loadProxyStatus();
      this.renderPage();
      alert(this.t('success'));
    } else {
      alert(data.error || this.t('error'));
    }
  }

  async stopProxy() {
    const res = await fetch('/api/proxy/stop', {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${this.token}` }
    });
    const data = await res.json();
    if (res.ok) {
      await this.loadProxyStatus();
      this.renderPage();
    } else {
      alert(data.error || this.t('error'));
    }
  }

  async testAllLatency() {
    const res = await fetch('/api/proxy/test', {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${this.token}` }
    });
    const data = await res.json();
    if (res.ok) {
      data.forEach(item => {
        const el = document.getElementById(`latency-${item.name}`);
        if (el) {
          el.textContent = item.error ? '×' : `${item.latency}ms`;
          el.className = item.error ? 'latency-bad' : 'latency-good';
        }
      });
    }
  }

  async selectNode(nodeName) {
    const res = await fetch('/api/proxy/start', {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${this.token}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ node_name: nodeName })
    });
    const data = await res.json();
    if (res.ok) {
      await this.loadProxyStatus();
      this.navigate('overview');
    } else {
      alert(data.error || this.t('error'));
    }
  }

  async refreshSub(name) {
    const res = await fetch(`/api/subscriptions/${name}/refresh`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${this.token}` }
    });
    if (res.ok) {
      await this.loadConfig();
      this.renderPage();
    } else {
      const data = await res.json();
      alert(data.error || this.t('error'));
    }
  }

  async deleteSub(name) {
    if (!confirm('Delete this subscription?')) return;
    const res = await fetch(`/api/subscriptions/${name}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${this.token}` }
    });
    if (res.ok) {
      await this.loadConfig();
      this.renderPage();
    }
  }

  showAddSubModal() {
    const name = prompt('Subscription name:');
    if (!name) return;
    const url = prompt('Subscription URL:');
    if (!url) return;
    fetch('/api/subscriptions', {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${this.token}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, url })
    }).then(async res => {
      if (res.ok) {
        await this.loadConfig();
        this.renderPage();
      }
    });
  }

  showAddNodeModal() {
    const link = prompt('Node link (vmess:// / vless:// / trojan:// / ss:// / anytls://):');
    if (!link) return;
    fetch('/api/nodes', {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${this.token}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ link })
    }).then(async res => {
      if (res.ok) {
        await this.loadConfig();
        this.renderPage();
      } else {
        const data = await res.json();
        alert(data.error || this.t('error'));
      }
    });
  }

  changeRouteMode(mode) {
    // Route mode is saved when starting proxy
    this.config.route_mode = mode;
  }

  toggleLang() {
    this.lang = this.lang === 'zh' ? 'en' : 'zh';
    localStorage.setItem('lang', this.lang);
    this.renderPage();
    this.renderSidebar();
  }

  logout() {
    localStorage.removeItem('token');
    this.token = null;
    location.reload();
  }
}

const app = new App();
