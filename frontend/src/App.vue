<template>
  <div class="app-container">
    <!-- Toast通知 -->
    <div class="toast-container">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        class="toast"
        :class="toast.intent"
      >
        <div class="toast-title">{{ toast.title }}</div>
        <div v-if="toast.message" class="toast-message">{{ toast.message }}</div>
      </div>
    </div>

    <!-- 头部 -->
    <div class="header">
      <span class="header-title">Windows 服务管理器</span>
      <div class="header-actions">
        <span v-if="!adminPrivileges" class="badge warning">非管理员模式</span>
        <button class="win11-button subtle" @click="openEnvDialog">
          <span class="icon">🏢</span>
          系统变量
        </button>
        <button class="win11-button subtle" @click="openSettingsDialog">
          <span class="icon">⚙️</span>
          应用设置
        </button>
        <button class="win11-button primary" @click="openAddDialog">
          <span class="icon">➕</span>
          添加服务
        </button>
      </div>
    </div>

    <!-- 主内容区 -->
    <div class="main-content">
      <div class="content-area">
        <div class="content-header">
          <span class="content-title">服务列表</span>
          <button class="win11-button subtle" @click="loadServices">
            <span class="icon">🔄</span>
            刷新
          </button>
        </div>

        <!-- 空状态 -->
        <div v-if="services.length === 0" class="empty-state">
          <div class="empty-state-icon">⚙️</div>
          <div class="empty-state-text">
            暂无服务<br />
            点击右上角"添加服务"按钮开始创建服务
          </div>
        </div>

        <!-- 服务列表表格 -->
        <table v-else class="win11-table">
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
            <tr
              v-for="service in services"
              :key="service.id"
              class="win11-table-row"
            >
              <td>
                <div class="service-name">{{ service.name }}</div>
                <div class="service-pid">PID: {{ service.pid || 'N/A' }}</div>
              </td>
              <td>
                <div class="service-status" :class="service.status">
                  <span class="status-dot"></span>
                  {{ getStatusText(service.status) }}
                </div>
              </td>
              <td>
                <div class="service-path">{{ service.exePath }}</div>
                <div v-if="service.args" class="service-args">
                  参数: {{ service.args }}
                </div>
              </td>
              <td>
                <label class="switch">
                  <input
                    type="checkbox"
                    :checked="service.autoStart"
                    @change="handleAutoStartToggle(service.id, $event.target.checked)"
                  />
                  <span class="slider"></span>
                </label>
              </td>
              <td>
                <div class="action-buttons">
                  <button
                    v-if="service.status === 'stopped'"
                    class="win11-button icon-button"
                    title="启动服务"
                    @click="handleStartService(service.id)"
                  >
                    ▶️
                  </button>
                  <button
                    v-else
                    class="win11-button icon-button secondary"
                    title="停止服务"
                    @click="handleStopService(service.id)"
                  >
                    ⏹️
                  </button>
                  <button
                    class="win11-button icon-button"
                    title="查看日志"
                    @click="handleViewLogs(service.id, service.name)"
                  >
                    📄
                  </button>
                  <button
                    class="win11-button icon-button delete"
                    title="删除服务"
                    @click="handleDeleteService(service.id)"
                  >
                    🗑️
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 状态栏 -->
    <div class="status-bar">
      总计服务: {{ serviceStats.total }} |
      运行中: {{ serviceStats.running }} |
      已停止: {{ serviceStats.stopped }}
      {{ adminPrivileges ? ' | 管理员权限' : ' | 普通权限' }}
    </div>

    <!-- 添加服务对话框 -->
    <div v-if="isAddDialogOpen" class="dialog-overlay" @click.self="closeAddDialog">
      <div class="dialog win11-dialog">
        <div class="dialog-header">
          <h3>添加新服务</h3>
        </div>
        <div class="dialog-content">
          <div class="form-group">
            <label class="required">服务名称</label>
            <input
              v-model="newService.name"
              type="text"
              class="win11-input"
              placeholder="输入服务名称"
            />
          </div>
          <div class="form-group">
            <label class="required">可执行文件路径</label>
            <div class="input-with-button">
              <input
                v-model="newService.exePath"
                type="text"
                class="win11-input"
                placeholder="输入程序路径"
              />
              <button class="win11-button" @click="handleSelectFile">📄 选择</button>
            </div>
          </div>
          <div class="form-group">
            <label>启动参数</label>
            <input
              v-model="newService.args"
              type="text"
              class="win11-input"
              placeholder="输入启动参数（可选）"
            />
          </div>
          <div class="form-group">
            <label>工作目录</label>
            <div class="input-with-button">
              <input
                v-model="newService.workingDir"
                type="text"
                class="win11-input"
                placeholder="工作目录（留空使用程序目录）"
              />
              <button class="win11-button" @click="handleSelectDirectory">📁 选择</button>
            </div>
          </div>
          <div class="form-group">
            <label>服务启动</label>
            <div class="hint-box">
              💡 服务创建后将自动启动并运行
            </div>
          </div>
        </div>
        <div class="dialog-actions">
          <button class="win11-button" @click="closeAddDialog">取消</button>
          <button class="win11-button primary" @click="handleCreateService">创建服务</button>
        </div>
      </div>
    </div>

    <!-- 权限警告对话框 -->
    <div v-if="showAdminWarning" class="dialog-overlay">
      <div class="dialog win11-dialog">
        <div class="dialog-header">
          <h3>权限警告</h3>
        </div>
        <div class="dialog-content center">
          <div class="warning-text">当前没有管理员权限，无法使用服务管理功能！</div>
          <div>请使用管理员权限重新启动程序以获得完整功能。</div>
        </div>
        <div class="dialog-actions">
          <button class="win11-button primary" @click="handleRestartAsAdmin">以管理员身份重启</button>
          <button class="win11-button" @click="showAdminWarning = false">暂时忽略</button>
        </div>
      </div>
    </div>

    <!-- 设置对话框 -->
    <div v-if="isSettingsDialogOpen" class="dialog-overlay" @click.self="closeSettingsDialog">
      <div class="dialog win11-dialog">
        <div class="dialog-header">
          <h3>应用设置</h3>
        </div>
        <div class="dialog-content">
          <div class="settings-section">
            <label>权限管理</label>
            <div class="settings-item">
              <span>当前权限状态</span>
              <span :class="['badge', adminPrivileges ? 'success' : 'warning']">
                {{ adminPrivileges ? '管理员权限' : '普通权限' }}
              </span>
            </div>
            <button v-if="!adminPrivileges" class="win11-button primary small" @click="handleRestartAsAdmin">
              以管理员身份重启
            </button>
          </div>
          <div class="settings-section">
            <label>开机自启动</label>
            <div class="settings-item toggle">
              <span>为此程序添加开机自启动项</span>
              <label class="switch">
                <input
                  type="checkbox"
                  :checked="autoStart"
                  @change="handleAppAutoStartToggle($event.target.checked)"
                />
                <span class="slider"></span>
              </label>
            </div>
          </div>
          <div class="settings-section">
            <label>应用信息</label>
            <div class="info-box">
              <div class="info-title">Windows Service Manager</div>
              <div class="info-text">现代化 Windows 服务管理工具</div>
              <div class="info-text">使程序以后台服务的形式运行</div>
              <div class="info-text">项目地址: https://github.com/sky22333/services</div>
            </div>
          </div>
        </div>
        <div class="dialog-actions">
          <button class="win11-button primary" @click="closeSettingsDialog">关闭</button>
        </div>
      </div>
    </div>

    <!-- 系统变量对话框 -->
    <div v-if="isEnvDialogOpen" class="dialog-overlay" @click.self="closeEnvDialog">
      <div class="dialog win11-dialog">
        <div class="dialog-header">
          <h3>添加系统环境变量</h3>
        </div>
        <div class="dialog-content">
          <div class="form-group">
            <label class="required">文件或目录路径</label>
            <div class="env-path-section">
              <div class="hint-text">💡 输入或选择要添加到系统PATH的文件/目录路径</div>
              <div class="input-with-buttons">
                <input
                  v-model="envPath"
                  type="text"
                  class="win11-input"
                  placeholder="例如: C:\Program Files\MyApp\bin"
                />
                <button class="win11-button small" @click="handleSelectEnvFile">📄 文件</button>
                <button class="win11-button small secondary" @click="handleSelectEnvDirectory">📁 目录</button>
              </div>
              <div class="env-info">
                <div><strong>功能介绍：</strong></div>
                <div><strong>说明：</strong>方便快捷的将程序添加到系统变量</div>
                <div><strong>使用：</strong>支持手动输入路径，支持选择程序或者选择目录</div>
                <div><strong>效果：</strong>快速将路径将添加到系统级PATH，重新打开终端即可使用</div>
              </div>
            </div>
          </div>
          <div class="form-group">
            <label>快捷操作</label>
            <button class="win11-button secondary" @click="handleOpenSystemEnvironmentSettings">
              打开系统环境变量设置界面
            </button>
          </div>
        </div>
        <div class="dialog-actions">
          <button class="win11-button" @click="closeEnvDialog">取消</button>
          <button
            class="win11-button primary"
            :disabled="!envPath.trim() || isAddingEnv"
            @click="handleAddEnvironmentVariable"
          >
            {{ isAddingEnv ? '添加中...' : '添加到PATH' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 删除确认对话框 -->
    <div v-if="isDeleteDialogOpen" class="dialog-overlay" @click.self="closeDeleteDialog">
      <div class="dialog">
        <div class="dialog-header">
          <h3>确认删除服务</h3>
        </div>
        <div class="dialog-content">
          <p>确定要删除服务 "{{ serviceToDelete?.name }}" 吗？</p>
          <p class="warning-text">服务将被删除！</p>
        </div>
        <div class="dialog-actions">
          <button class="win11-button" @click="closeDeleteDialog">取消</button>
          <button class="win11-button danger" @click="confirmDeleteService">删除</button>
        </div>
      </div>
    </div>

    <!-- 日志查看对话框 -->
    <div v-if="isLogsDialogOpen" class="dialog-overlay" @click.self="closeLogsDialog">
      <div class="dialog logs-dialog">
        <div class="dialog-header">
          <h3>服务日志 - {{ serviceToViewLogs?.name }}</h3>
        </div>
        <div class="dialog-content">
          <div class="logs-container">
            <pre>{{ serviceLogs || '暂无日志内容' }}</pre>
          </div>
        </div>
        <div class="dialog-actions">
          <button class="win11-button" @click="handleOpenLogsDirectory">📁 打开日志目录</button>
          <button class="win11-button" @click="handleCopyLogsPath">📋 复制路径</button>
          <button class="win11-button primary" @click="closeLogsDialog">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
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
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime'

// 状态
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
const newService = ref({
  name: '',
  exePath: '',
  args: '',
  workingDir: ''
})
const toasts = ref([])
let toastId = 0

// 计算属性
const serviceStats = computed(() => ({
  total: services.value.length,
  running: services.value.filter(s => s.status === 'running').length,
  stopped: services.value.filter(s => s.status === 'stopped').length
}))

// 方法
const showToast = (title, message, intent = 'success') => {
  const id = ++toastId
  toasts.value.push({ id, title, message, intent })
  setTimeout(() => {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }, 3000)
}

const getStatusText = (status) => {
  const statusMap = {
    running: '运行中',
    stopped: '已停止',
    error: '错误'
  }
  return statusMap[status] || status
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
    closeAddDialog()
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
    closeDeleteDialog()
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
    envPath.value = ''
  } catch (error) {
    console.error('环境变量添加失败:', error)

    if (error.toString().includes('Access is denied') ||
        error.toString().includes('access denied') ||
        error.toString().includes('无法读取现有PATH变量')) {
      try {
        const diagnosis = await DiagnoseEnvironmentAccess()
        console.log('权限诊断结果:', diagnosis)

        let errorMsg = '权限不足，无法修改系统环境变量。\n\n'
        if (!diagnosis.registry_full) errorMsg += '• 注册表完整权限: 失败\n'
        if (!diagnosis.registry_write) errorMsg += '• 注册表写入权限: 失败\n'
        if (!diagnosis.path_read) errorMsg += '• PATH变量读取: 失败\n'
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
      showToast('成功', '日志路径已复制到剪贴板')
    } else {
      showToast('错误', '日志文件不存在', 'error')
    }
  } catch (error) {
    showToast('错误', '复制日志路径失败: ' + error, 'error')
  }
}

// 对话框控制
const openAddDialog = () => { isAddDialogOpen.value = true }
const closeAddDialog = () => { isAddDialogOpen.value = false }
const openSettingsDialog = () => { isSettingsDialogOpen.value = true }
const closeSettingsDialog = () => { isSettingsDialogOpen.value = false }
const openEnvDialog = () => { isEnvDialogOpen.value = true }
const closeEnvDialog = () => {
  isEnvDialogOpen.value = false
  envPath.value = ''
}
const closeDeleteDialog = () => {
  isDeleteDialogOpen.value = false
  serviceToDelete.value = null
}
const closeLogsDialog = () => {
  isLogsDialogOpen.value = false
  serviceToViewLogs.value = null
  serviceLogs.value = ''
}

// 生命周期
let unsubscribeStatusChanged = null
let unsubscribeServicesUpdated = null

onMounted(() => {
  loadServices()
  checkAdminRights()
  checkAutoStartStatus()

  unsubscribeStatusChanged = EventsOn('service-status-changed', (data) => {
    services.value = services.value.map(service =>
      service.id === data.serviceId
        ? { ...service, status: data.status, pid: data.pid }
        : service
    )
  })

  unsubscribeServicesUpdated = EventsOn('services-updated', (serviceList) => {
    services.value = serviceList || []
  })
})

onUnmounted(() => {
  if (unsubscribeStatusChanged) EventsOff('service-status-changed')
  if (unsubscribeServicesUpdated) EventsOff('services-updated')
})
</script>

<style scoped>
/* 组件特定样式 */
</style>
