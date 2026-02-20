<template>
  <div class="app-container">
    <div class="toast-container">
      <div v-for="toast in toasts" :key="toast.id" :class="['toast', toast.type]">
        <div class="toast-title">{{ toast.title }}</div>
        <div v-if="toast.message" class="toast-message">{{ toast.message }}</div>
      </div>
    </div>

    <div class="header">
      <span class="header-title">Windows 服务管理器</span>
      <div class="header-actions">
        <span v-if="!adminPrivileges" class="badge badge-warning">非管理员模式</span>
        <button class="btn btn-subtle" @click="isEnvDialogOpen = true">
          <span class="icon">🏢</span>
          系统变量
        </button>
        <button class="btn btn-subtle" @click="isSettingsDialogOpen = true">
          <span class="icon">⚙️</span>
          应用设置
        </button>
        <button class="btn btn-primary" @click="isAddDialogOpen = true">
          <span class="icon">➕</span>
          添加服务
        </button>
      </div>
    </div>

    <div class="main-content">
      <div class="content-area">
        <div class="content-header">
          <span class="content-title">服务列表</span>
          <button class="btn btn-subtle" @click="loadServices">
            <span class="icon">🔄</span>
            刷新
          </button>
        </div>

        <div v-if="services.length === 0" class="empty-state">
          <div class="empty-state-icon">⚙️</div>
          <div class="empty-state-text">
            暂无服务<br />
            点击右上角"添加服务"按钮开始创建服务
          </div>
        </div>

        <table v-else class="win11-table slide-in">
          <thead class="win11-table-header">
            <tr>
              <th>服务名称</th>
              <th>状态</th>
              <th>程序路径</th>
              <th>开机自启</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="service in services" :key="service.id" class="win11-table-row">
              <td>
                <div class="service-name">{{ service.name }}</div>
                <div class="service-pid">PID: {{ service.pid || 'N/A' }}</div>
              </td>
              <td>
                <div :class="['service-status', service.status]">
                  <span :class="['status-dot', service.status]"></span>
                  {{ getStatusText(service.status) }}
                </div>
              </td>
              <td>
                <div class="exe-path">{{ service.exePath }}</div>
                <div v-if="service.args" class="service-args">参数: {{ service.args }}</div>
              </td>
              <td>
                <label class="switch">
                  <input type="checkbox" :checked="service.autoStart || false" @change="handleAutoStartToggle(service.id, $event.target.checked)" />
                  <span class="slider"></span>
                </label>
              </td>
              <td>
                <div class="action-buttons">
                  <button v-if="service.status === 'stopped'" class="btn btn-icon" title="启动服务" @click="handleStartService(service.id)">
                    <span class="icon">▶️</span>
                  </button>
                  <button v-else class="btn btn-icon btn-secondary" title="停止服务" @click="handleStopService(service.id)">
                    <span class="icon">⏹️</span>
                  </button>
                  <button class="btn btn-icon" title="查看日志" @click="handleViewLogs(service.id, service.name)">
                    <span class="icon">📄</span>
                  </button>
                  <button class="btn btn-icon btn-delete" title="删除服务" @click="handleDeleteService(service.id)">
                    <span class="icon">🗑️</span>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="status-bar">
      <span>总计服务: {{ serviceStats.total }} | 运行中: {{ serviceStats.running }} | 已停止: {{ serviceStats.stopped }}{{ adminPrivileges ? ' | 管理员权限' : ' | 普通权限' }}</span>
    </div>

    <!-- 添加服务对话框 -->
    <div v-if="isAddDialogOpen" class="dialog-overlay" @click.self="isAddDialogOpen = false">
      <div class="dialog win11-dialog">
        <div class="dialog-header">
          <h2>添加新服务</h2>
        </div>
        <div class="dialog-body">
          <div class="form-group">
            <label>服务名称 <span class="required">*</span></label>
            <input type="text" v-model="newService.name" placeholder="输入服务名称" class="win11-input" />
          </div>
          <div class="form-group">
            <label>可执行文件路径 <span class="required">*</span></label>
            <div class="input-with-button">
              <input type="text" v-model="newService.exePath" placeholder="输入程序路径" class="win11-input" />
              <button class="btn" @click="handleSelectFile">
                <span class="icon">📄</span>
                选择
              </button>
            </div>
          </div>
          <div class="form-group">
            <label>启动参数</label>
            <input type="text" v-model="newService.args" placeholder="输入启动参数（可选）" class="win11-input" />
          </div>
          <div class="form-group">
            <label>工作目录</label>
            <div class="input-with-button">
              <input type="text" v-model="newService.workingDir" placeholder="工作目录（留空使用程序目录）" class="win11-input" />
              <button class="btn" @click="handleSelectDirectory">
                <span class="icon">📁</span>
                选择
              </button>
            </div>
          </div>
          <div class="form-tip">
            💡 服务创建后将自动启动并运行
          </div>
        </div>
        <div class="dialog-footer">
          <button class="btn btn-secondary" @click="isAddDialogOpen = false">取消</button>
          <button class="btn btn-primary" @click="handleCreateService">创建服务</button>
        </div>
      </div>
    </div>

    <!-- 权限警告对话框 -->
    <div v-if="showAdminWarning" class="dialog-overlay">
      <div class="dialog win11-dialog">
        <div class="dialog-header">
          <h2>权限警告</h2>
        </div>
        <div class="dialog-body">
          <div class="warning-text">
            <p class="warning-main">当前没有管理员权限，无法使用服务管理功能！</p>
            <p>请使用管理员权限重新启动程序以获得完整功能。</p>
          </div>
        </div>
        <div class="dialog-footer">
          <button class="btn btn-primary" @click="handleRestartAsAdmin">以管理员身份重启</button>
          <button class="btn btn-secondary" @click="showAdminWarning = false">暂时忽略</button>
        </div>
      </div>
    </div>

    <!-- 设置对话框 -->
    <div v-if="isSettingsDialogOpen" class="dialog-overlay" @click.self="isSettingsDialogOpen = false">
      <div class="dialog win11-dialog">
        <div class="dialog-header">
          <h2>应用设置</h2>
        </div>
        <div class="dialog-body">
          <div class="settings-section">
            <h3>权限管理</h3>
            <div class="settings-item">
              <span>当前权限状态</span>
              <span :class="['badge', adminPrivileges ? 'badge-success' : 'badge-warning']">
                {{ adminPrivileges ? '管理员权限' : '普通权限' }}
              </span>
            </div>
            <button v-if="!adminPrivileges" class="btn btn-primary btn-small" @click="handleRestartAsAdmin">
              以管理员身份重启
            </button>
          </div>
          <div class="settings-section">
            <h3>开机自启动</h3>
            <div class="settings-card">
              <div class="settings-item">
                <span>为此程序添加开机自启动项</span>
                <label class="switch">
                  <input type="checkbox" :checked="autoStart" @change="handleAppAutoStartToggle($event.target.checked)" />
                  <span class="slider"></span>
                </label>
              </div>
            </div>
          </div>
          <div class="settings-section">
            <h3>应用信息</h3>
            <div class="settings-card">
              <p class="app-name">Windows Service Manager</p>
              <p class="app-desc">现代化 Windows 服务管理工具</p>
              <p class="app-desc">使程序以后台服务的形式运行</p>
              <p class="app-desc">项目地址: https://github.com/sky22333/services</p>
            </div>
          </div>
        </div>
        <div class="dialog-footer">
          <button class="btn btn-primary" @click="isSettingsDialogOpen = false">关闭</button>
        </div>
      </div>
    </div>

    <!-- 系统变量对话框 -->
    <div v-if="isEnvDialogOpen" class="dialog-overlay" @click.self="closeEnvDialog">
      <div class="dialog win11-dialog">
        <div class="dialog-header">
          <h2>添加系统环境变量</h2>
        </div>
        <div class="dialog-body">
          <div class="env-section">
            <h3>文件或目录路径 <span class="required">*</span></h3>
            <div class="env-card">
              <p class="env-tip">💡 输入或选择要添加到系统PATH的文件/目录路径</p>
              <div class="input-with-buttons">
                <input type="text" v-model="envPath" placeholder="例如: C:\Program Files\MyApp\bin" class="win11-input" />
                <div class="button-group">
                  <button class="btn btn-small" title="选择可执行文件（自动提取目录）" @click="handleSelectEnvFile">
                    <span class="icon">📄</span>
                    文件
                  </button>
                  <button class="btn btn-small btn-secondary" title="直接选择目录" @click="handleSelectEnvDirectory">
                    <span class="icon">📁</span>
                    目录
                  </button>
                </div>
              </div>
              <div class="env-info">
                <p><strong>功能介绍：</strong></p>
                <p><strong>说明：</strong>方便快捷的将程序添加到系统变量</p>
                <p><strong>使用：</strong>支持手动输入路径，支持选择程序或者选择目录</p>
                <p><strong>效果：</strong>快速将路径将添加到系统级PATH，重新打开终端即可使用</p>
              </div>
            </div>
          </div>
          <div class="env-section">
            <h3>快捷操作</h3>
            <div class="settings-card">
              <button class="btn btn-secondary btn-small" @click="handleOpenSystemEnvironmentSettings">
                打开系统环境变量设置界面
              </button>
            </div>
          </div>
        </div>
        <div class="dialog-footer">
          <button class="btn btn-secondary" @click="closeEnvDialog">取消</button>
          <button class="btn btn-primary" @click="handleAddEnvironmentVariable" :disabled="!envPath.trim() || isAddingEnv">
            {{ isAddingEnv ? '添加中...' : '添加到PATH' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 删除确认对话框 -->
    <div v-if="isDeleteDialogOpen" class="dialog-overlay" @click.self="isDeleteDialogOpen = false">
      <div class="dialog">
        <div class="dialog-header">
          <h2>确认删除服务</h2>
        </div>
        <div class="dialog-body">
          <p>确定要删除服务 "{{ serviceToDelete?.name }}" 吗？</p>
          <p class="delete-warning">服务将被删除！</p>
        </div>
        <div class="dialog-footer">
          <button class="btn btn-secondary" @click="isDeleteDialogOpen = false">取消</button>
          <button class="btn btn-danger" @click="confirmDeleteService">删除</button>
        </div>
      </div>
    </div>

    <!-- 日志查看对话框 -->
    <div v-if="isLogsDialogOpen" class="dialog-overlay" @click.self="isLogsDialogOpen = false">
      <div class="dialog dialog-large win11-dialog">
        <div class="dialog-header">
          <h2>服务日志 - {{ serviceToViewLogs?.name }}</h2>
        </div>
        <div class="dialog-body">
          <div class="logs-container">
            {{ serviceLogs || '暂无日志' }}
          </div>
        </div>
        <div class="dialog-footer">
          <button class="btn btn-secondary btn-small" @click="handleOpenLogsDirectory">
            <span class="icon">📁</span>
            打开日志目录
          </button>
          <button class="btn btn-secondary btn-small" @click="handleCopyLogsPath">
            <span class="icon">📋</span>
            复制路径
          </button>
          <button class="btn btn-primary btn-small" @click="isLogsDialogOpen = false">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime'
import { 
  GetServices, 
  CreateService, 
  StartService, 
  StopService, 
  DeleteService,
  SelectFile,
  SelectDirectory,
  CheckAdminPrivileges,
  SetAutoStart,
  GetAutoStartStatus,
  SetServiceAutoStart,
  RestartAsAdmin,
  AddPathVariable,
  OpenSystemEnvironmentSettings,
  ValidatePathExists,
  DiagnoseEnvironmentAccess,
  GetServiceLogs,
  GetServiceLogsPath,
  OpenLogsDirectory
} from "../wailsjs/go/main/App"
import './App.css'

const services = ref([])
const isAddDialogOpen = ref(false)
const isSettingsDialogOpen = ref(false)
const isDeleteDialogOpen = ref(false)
const isEnvDialogOpen = ref(false)
const isLogsDialogOpen = ref(false)
const serviceToDelete = ref(null)
const serviceToViewLogs = ref(null)
const serviceLogs = ref('')
const adminPrivileges = ref(false)
const autoStart = ref(false)
const showAdminWarning = ref(false)
const envPath = ref('')
const isAddingEnv = ref(false)
const toasts = ref([])
let toastId = 0

const newService = ref({
  name: '',
  exePath: '',
  args: '',
  workingDir: ''
})

const serviceStats = computed(() => ({
  total: services.value.length,
  running: services.value.filter(s => s.status === 'running').length,
  stopped: services.value.filter(s => s.status === 'stopped').length
}))

const showToast = (title, message, type = 'success') => {
  const id = ++toastId
  toasts.value.push({ id, title, message, type })
  setTimeout(() => {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }, 3000)
}

const getStatusText = (status) => {
  return status === 'running' ? '运行中' : status === 'error' ? '错误' : '已停止'
}

const loadServices = async () => {
  try {
    const serviceList = await GetServices()
    services.value = serviceList || []
  } catch (error) {
    showToast('错误', '加载服务列表失败: ' + error, 'error')
  }
}

const checkAdminRights = async () => {
  try {
    const isAdmin = await CheckAdminPrivileges()
    adminPrivileges.value = isAdmin
    if (!isAdmin) {
      showAdminWarning.value = true
    }
  } catch (error) {
    console.error('检查权限失败:', error)
  }
}

const checkAutoStartStatus = async () => {
  try {
    const status = await GetAutoStartStatus()
    autoStart.value = status
  } catch (error) {
    console.error('检查开机自启状态失败:', error)
  }
}

const handleAppAutoStartToggle = async (enabled) => {
  try {
    await SetAutoStart(enabled)
    autoStart.value = enabled
    showToast('成功', `开机自启动已${enabled ? '启用' : '禁用'}`)
  } catch (error) {
    showToast('错误', '设置开机自启动失败: ' + error, 'error')
  }
}

const handleRestartAsAdmin = async () => {
  try {
    await RestartAsAdmin()
  } catch (error) {
    showToast('错误', '以管理员身份重启失败: ' + error, 'error')
  }
}

const handleCreateService = async () => {
  if (!newService.value.name || !newService.value.exePath) {
    showToast('验证错误', '请填写服务名称和可执行文件路径', 'error')
    return
  }

  try {
    await CreateService(newService.value)
    showToast('成功', '服务创建成功')
    isAddDialogOpen.value = false
    newService.value = { name: '', exePath: '', args: '', workingDir: '' }
    loadServices()
  } catch (error) {
    showToast('错误', '创建服务失败: ' + error, 'error')
  }
}

const handleStartService = async (serviceId) => {
  try {
    await StartService(serviceId)
    showToast('成功', '服务启动成功')
    loadServices()
  } catch (error) {
    showToast('错误', '启动服务失败: ' + error, 'error')
  }
}

const handleStopService = async (serviceId) => {
  try {
    await StopService(serviceId)
    showToast('成功', '服务停止成功')
    loadServices()
  } catch (error) {
    showToast('错误', '停止服务失败: ' + error, 'error')
  }
}

const handleDeleteService = (serviceId) => {
  const service = services.value.find(s => s.id === serviceId)
  serviceToDelete.value = service
  isDeleteDialogOpen.value = true
}

const confirmDeleteService = async () => {
  if (!serviceToDelete.value) return
  
  try {
    await DeleteService(serviceToDelete.value.id)
    showToast('成功', '服务删除成功')
    loadServices()
  } catch (error) {
    showToast('错误', '删除服务失败: ' + error, 'error')
  } finally {
    isDeleteDialogOpen.value = false
    serviceToDelete.value = null
  }
}

const handleAutoStartToggle = async (serviceId, enabled) => {
  try {
    await SetServiceAutoStart(serviceId, enabled)
    showToast('成功', enabled ? '已启用开机自启' : '已禁用开机自启')
    loadServices()
  } catch (error) {
    showToast('错误', '设置开机自启失败: ' + error, 'error')
  }
}

const handleSelectFile = async () => {
  try {
    const filePath = await SelectFile()
    if (filePath) {
      newService.value.exePath = filePath
    }
  } catch (error) {
    showToast('错误', '选择文件失败: ' + error, 'error')
  }
}

const handleSelectDirectory = async () => {
  try {
    const dirPath = await SelectDirectory()
    if (dirPath) {
      newService.value.workingDir = dirPath
    }
  } catch (error) {
    showToast('错误', '选择目录失败: ' + error, 'error')
  }
}

const handleSelectEnvFile = async () => {
  try {
    const filePath = await SelectFile()
    if (filePath) {
      envPath.value = filePath
    }
  } catch (error) {
    showToast('错误', '选择文件失败: ' + error, 'error')
  }
}

const handleSelectEnvDirectory = async () => {
  try {
    const dirPath = await SelectDirectory()
    if (dirPath) {
      envPath.value = dirPath
    }
  } catch (error) {
    showToast('错误', '选择目录失败: ' + error, 'error')
  }
}

const handleAddEnvironmentVariable = async () => {
  if (!envPath.value.trim()) {
    showToast('验证错误', '请输入或选择文件路径', 'error')
    return
  }

  isAddingEnv.value = true
  try {
    const exists = await ValidatePathExists(envPath.value)
    if (!exists) {
      showToast('验证错误', '指定的路径不存在', 'error')
      return
    }

    await AddPathVariable(envPath.value)
    showToast('成功', 'PATH环境变量添加成功！新打开的命令行窗口将生效')
    closeEnvDialog()
  } catch (error) {
    console.error('环境变量添加失败:', error)
    
    if (error.toString().includes('Access is denied') || 
        error.toString().includes('access denied') ||
        error.toString().includes('无法读取现有PATH变量')) {
      
      try {
        const diagnosis = await DiagnoseEnvironmentAccess()
        console.log('权限诊断结果:', diagnosis)
        
        let errorMsg = '权限不足，无法修改系统环境变量。\n\n'
        
        if (!diagnosis.registry_full) {
          errorMsg += '• 注册表完整权限: 失败\n'
        }
        if (!diagnosis.registry_write) {
          errorMsg += '• 注册表写入权限: 失败\n'
        }
        if (!diagnosis.path_read) {
          errorMsg += '• PATH变量读取: 失败\n'
        }
        
        errorMsg += '\n请确认：\n'
        errorMsg += '1. 程序以管理员身份运行\n'
        errorMsg += '2. 系统未被组策略限制环境变量修改\n'
        errorMsg += '3. 杀毒软件未阻止注册表访问'
        
        showToast('权限诊断', errorMsg, 'error')
      } catch (diagError) {
        showToast('错误', '添加环境变量失败: ' + error + '\n诊断失败: ' + diagError, 'error')
      }
    } else {
      showToast('错误', '添加环境变量失败: ' + error, 'error')
    }
  } finally {
    isAddingEnv.value = false
  }
}

const handleOpenSystemEnvironmentSettings = async () => {
  try {
    await OpenSystemEnvironmentSettings()
  } catch (error) {
    showToast('错误', '打开系统环境变量设置失败: ' + error, 'error')
  }
}

const handleViewLogs = async (serviceId, serviceName) => {
  try {
    serviceToViewLogs.value = { id: serviceId, name: serviceName }
    let logs = ''
    
    try {
      logs = await GetServiceLogs(serviceId)
    } catch (apiError) {
      logs = `获取日志失败: ${apiError.message}\n\n服务可能尚未启动，或者日志文件不存在。\n\n请尝试：\n1. 确保服务已启动\n2. 检查服务的可执行文件路径是否正确\n3. 重启服务管理器`
    }
    
    serviceLogs.value = logs
    isLogsDialogOpen.value = true
  } catch (error) {
    showToast('错误', '打开日志查看器失败: ' + error, 'error')
  }
}

const handleOpenLogsDirectory = async () => {
  try {
    await OpenLogsDirectory(serviceToViewLogs.value?.id)
  } catch (error) {
    showToast('错误', '打开日志目录失败: ' + error, 'error')
  }
}

const handleCopyLogsPath = async () => {
  try {
    const logPath = await GetServiceLogsPath(serviceToViewLogs.value?.id)
    if (logPath) {
      await navigator.clipboard.writeText(logPath)
      showToast('成功', '日志路径已复制到剪贴板', 'success')
    } else {
      showToast('错误', '日志文件不存在', 'error')
    }
  } catch (error) {
    showToast('错误', '复制日志路径失败: ' + error, 'error')
  }
}

const closeEnvDialog = () => {
  isEnvDialogOpen.value = false
  envPath.value = ''
}

onMounted(() => {
  loadServices()
  checkAdminRights()
  checkAutoStartStatus()
  
  EventsOn('service-status-changed', (data) => {
    services.value = services.value.map(service => 
      service.id === data.serviceId 
        ? { ...service, status: data.status, pid: data.pid }
        : service
    )
  })
  
  EventsOn('services-updated', (serviceList) => {
    services.value = serviceList || []
  })
})

onUnmounted(() => {
  EventsOff('service-status-changed')
  EventsOff('services-updated')
})
</script>

<style scoped>
.toast-container {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.toast {
  padding: 12px 20px;
  border-radius: 8px;
  background: white;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  animation: slideIn 0.3s ease;
}

.toast.success {
  border-left: 4px solid #107c10;
}

.toast.error {
  border-left: 4px solid #c42b1c;
}

.toast-title {
  font-weight: 600;
  margin-bottom: 4px;
}

.toast-message {
  font-size: 13px;
  color: #666;
  white-space: pre-wrap;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateX(20px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.header-title {
  font-size: 16px;
  font-weight: 600;
}

.header-actions {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 10px;
}

.btn {
  padding: 8px 16px;
  border-radius: 8px;
  border: none;
  cursor: pointer;
  font-size: 14px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: all 0.1s ease;
  background: rgba(0, 0, 0, 0.05);
}

.btn:hover {
  transform: scale(1.02);
}

.btn:active {
  transform: scale(0.98);
}

.btn-primary {
  background: #0078d4;
  color: white;
}

.btn-primary:hover {
  background: #106ebe;
}

.btn-secondary {
  background: rgba(0, 0, 0, 0.08);
}

.btn-subtle {
  background: transparent;
}

.btn-subtle:hover {
  background: rgba(0, 0, 0, 0.05);
}

.btn-danger {
  background: #d13438;
  color: white;
}

.btn-danger:hover {
  background: #a4262c;
}

.btn-small {
  padding: 6px 12px;
  font-size: 13px;
}

.btn-icon {
  padding: 6px;
  background: transparent;
}

.btn-icon:hover {
  background: rgba(0, 0, 0, 0.05);
}

.btn-delete:hover {
  background: #fef2f2;
  color: #dc2626;
}

.icon {
  font-size: 16px;
}

.badge {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.badge-success {
  background: rgba(16, 124, 16, 0.1);
  color: #107c10;
}

.badge-warning {
  background: rgba(255, 186, 0, 0.1);
  color: #ffb000;
}

.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.content-title {
  font-size: 14px;
  font-weight: 600;
}

.service-name {
  font-weight: 600;
  font-size: 14px;
}

.service-pid {
  font-size: 12px;
  color: #666;
  margin-top: 4px;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.status-dot.running {
  background: #107c10;
}

.status-dot.stopped {
  background: #605e5c;
}

.status-dot.error {
  background: #c42b1c;
}

.exe-path {
  font-size: 13px;
  word-break: break-all;
}

.service-args {
  font-size: 11px;
  color: #666;
  font-style: italic;
  margin-top: 4px;
}

.action-buttons {
  display: flex;
  gap: 6px;
  align-items: center;
}

.switch {
  position: relative;
  display: inline-block;
  width: 44px;
  height: 24px;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #ccc;
  transition: 0.3s;
  border-radius: 24px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.3s;
  border-radius: 50%;
}

input:checked + .slider {
  background-color: #0078d4;
}

input:checked + .slider:before {
  transform: translateX(20px);
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.dialog {
  background: white;
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
  max-width: 500px;
  width: 90%;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.dialog-large {
  max-width: 800px;
  width: 90%;
}

.dialog-header {
  padding: 20px 24px 12px;
  border-bottom: 1px solid #eee;
}

.dialog-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.dialog-body {
  padding: 20px 24px;
  overflow-y: auto;
  flex: 1;
  text-align: left;
}

.dialog-footer {
  padding: 16px 24px;
  border-top: 1px solid #eee;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  font-weight: 500;
  font-size: 14px;
}

.required {
  color: #d13438;
}

.win11-input {
  width: 100%;
  padding: 10px 12px;
  border-radius: 8px;
  border: 1px solid rgba(0, 0, 0, 0.1);
  background: rgba(255, 255, 255, 0.9);
  font-size: 14px;
  transition: all 0.2s ease;
}

.win11-input:focus {
  outline: none;
  border-color: #0078d4;
  box-shadow: 0 0 0 2px rgba(0, 120, 212, 0.2);
}

.input-with-button {
  display: flex;
  gap: 8px;
}

.input-with-button .win11-input {
  flex: 1;
}

.input-with-buttons {
  display: flex;
  gap: 8px;
  align-items: center;
}

.input-with-buttons .win11-input {
  flex: 1;
}

.button-group {
  display: flex;
  gap: 4px;
}

.form-tip {
  padding: 10px 12px;
  background: #f3f4f6;
  border-radius: 6px;
  border: 1px solid #e5e7eb;
  font-size: 13px;
  color: #666;
  font-style: italic;
}

.warning-text {
  text-align: center;
}

.warning-main {
  font-size: 16px;
  font-weight: 600;
  color: #d13438;
  margin-bottom: 12px;
}

.settings-section {
  margin-bottom: 20px;
}

.settings-section h3 {
  margin: 0 0 12px;
  font-size: 14px;
  font-weight: 600;
}

.settings-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.settings-card {
  padding: 16px;
  background: rgba(255, 255, 255, 0.5);
  border-radius: 12px;
  backdrop-filter: blur(10px);
}

.app-name {
  font-size: 15px;
  font-weight: 600;
  margin: 0 0 8px;
}

.app-desc {
  font-size: 13px;
  color: #666;
  margin: 4px 0;
}

.env-section {
  margin-bottom: 20px;
}

.env-section h3 {
  margin: 0 0 12px;
  font-size: 14px;
  font-weight: 600;
}

.env-card {
  padding: 16px;
  background: rgba(255, 255, 255, 0.5);
  border-radius: 12px;
  backdrop-filter: blur(10px);
  border: 1px solid #e5e7eb;
}

.env-tip {
  color: #666;
  margin-bottom: 12px;
}

.env-info {
  margin-top: 12px;
  padding: 8px 12px;
  background: #f8f9fa;
  border-radius: 6px;
  border: 1px solid #e9ecef;
  font-size: 12px;
  color: #666;
}

.env-info p {
  margin: 4px 0;
}

.delete-warning {
  color: #d13438;
  margin-top: 8px;
}

.logs-container {
  background: #f5f5f5;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  padding: 12px;
  max-height: 500px;
  min-height: 300px;
  overflow: auto;
  font-family: Consolas, Monaco, monospace;
  font-size: 13px;
  white-space: pre-wrap;
  word-break: break-word;
  color: #333;
  text-align: left;
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
