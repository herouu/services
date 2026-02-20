<template>
  <div class="app-container">
    <div class="header">
      <span class="header-title">Windows 服务管理器</span>
      <div class="header-actions">
        <span v-if="!adminPrivileges" class="badge badge-warning">非管理员模式</span>
        <button class="win11-button" @click="isEnvDialogOpen = true">
          <span class="icon">🏢</span> 系统变量
        </button>
        <button class="win11-button" @click="isSettingsDialogOpen = true">
          <span class="icon">⚙️</span> 应用设置
        </button>
        <button class="win11-button win11-button-primary" @click="isAddDialogOpen = true">
          <span class="icon">➕</span> 添加服务
        </button>
      </div>
    </div>

    <!-- 添加服务对话框 -->
    <div v-if="isAddDialogOpen" class="dialog-overlay" @click.self="isAddDialogOpen = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3>添加新服务</h3>
          <button class="dialog-close" @click="isAddDialogOpen = false">✕</button>
        </div>
        <div class="dialog-body">
          <div class="form-group">
            <label>服务名称 <span class="required">*</span></label>
            <input 
              type="text" 
              v-model="newService.name" 
              placeholder="输入服务名称"
              class="win11-input"
            />
          </div>
          
          <div class="form-group">
            <label>可执行文件路径 <span class="required">*</span></label>
            <div class="input-with-button">
              <input 
                type="text" 
                v-model="newService.exePath" 
                placeholder="输入程序路径"
                class="win11-input"
              />
              <button class="win11-button" @click="handleSelectFile">选择</button>
            </div>
          </div>
          
          <div class="form-group">
            <label>启动参数</label>
            <input 
              type="text" 
              v-model="newService.args" 
              placeholder="输入启动参数（可选）"
              class="win11-input"
            />
          </div>
          
          <div class="form-group">
            <label>工作目录</label>
            <div class="input-with-button">
              <input 
                type="text" 
                v-model="newService.workingDir" 
                placeholder="工作目录（留空使用程序目录）"
                class="win11-input"
              />
              <button class="win11-button" @click="handleSelectDirectory">选择</button>
            </div>
          </div>
          
          <div class="form-group">
            <label>服务启动</label>
            <div class="info-box">
              💡 服务创建后将自动启动并运行
            </div>
          </div>
        </div>
        <div class="dialog-footer">
          <button class="win11-button" @click="isAddDialogOpen = false">取消</button>
          <button class="win11-button win11-button-primary" @click="handleCreateService">创建服务</button>
        </div>
      </div>
    </div>

    <!-- 权限警告对话框 -->
    <div v-if="showAdminWarning" class="dialog-overlay">
      <div class="dialog">
        <div class="dialog-header">
          <h3>权限警告</h3>
        </div>
        <div class="dialog-body">
          <div class="warning-content">
            <p class="warning-text">当前没有管理员权限，无法使用服务管理功能！</p>
            <p>请使用管理员权限重新启动程序以获得完整功能。</p>
          </div>
        </div>
        <div class="dialog-footer">
          <button class="win11-button win11-button-primary" @click="handleRestartAsAdmin">以管理员身份重启</button>
          <button class="win11-button" @click="showAdminWarning = false">暂时忽略</button>
        </div>
      </div>
    </div>

    <!-- 设置对话框 -->
    <div v-if="isSettingsDialogOpen" class="dialog-overlay" @click.self="isSettingsDialogOpen = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3>应用设置</h3>
          <button class="dialog-close" @click="isSettingsDialogOpen = false">✕</button>
        </div>
        <div class="dialog-body">
          <div class="settings-section">
            <label>权限管理</label>
            <div class="settings-content">
              <div class="setting-row">
                <span>当前权限状态</span>
                <span :class="['badge', adminPrivileges ? 'badge-success' : 'badge-warning']">
                  {{ adminPrivileges ? '管理员权限' : '普通权限' }}
                </span>
              </div>
              <div v-if="!adminPrivileges" class="setting-action">
                <button class="win11-button" @click="handleRestartAsAdmin">以管理员身份重启</button>
              </div>
            </div>
          </div>

          <div class="settings-section">
            <label>开机自启动</label>
            <div class="settings-content">
              <div class="setting-row">
                <span>为此程序添加开机自启动项</span>
                <label class="switch">
                  <input type="checkbox" v-model="autoStart" @change="handleAppAutoStartToggle">
                  <span class="slider"></span>
                </label>
              </div>
            </div>
          </div>

          <div class="settings-section">
            <label>应用信息</label>
            <div class="settings-content app-info">
              <p class="app-title">Windows Service Manager</p>
              <p>现代化 Windows 服务管理工具</p>
              <p>使程序以后台服务的形式运行</p>
              <p>项目地址: https://github.com/sky22333/services</p>
            </div>
          </div>
        </div>
        <div class="dialog-footer">
          <button class="win11-button win11-button-primary" @click="isSettingsDialogOpen = false">关闭</button>
        </div>
      </div>
    </div>

    <!-- 系统变量对话框 -->
    <div v-if="isEnvDialogOpen" class="dialog-overlay" @click.self="isEnvDialogOpen = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3>添加系统环境变量</h3>
          <button class="dialog-close" @click="closeEnvDialog">✕</button>
        </div>
        <div class="dialog-body">
          <div class="form-group">
            <label>文件或目录路径 <span class="required">*</span></label>
            <div class="env-info-box">
              💡 输入或选择要添加到系统PATH的文件/目录路径
              
              <div class="input-with-button" style="margin-top: 12px;">
                <input 
                  type="text" 
                  v-model="envPath" 
                  placeholder="例如: C:\Program Files\MyApp\bin"
                  class="win11-input"
                />
                <div class="button-group">
                  <button class="win11-button" @click="handleSelectEnvFile">文件</button>
                  <button class="win11-button" @click="handleSelectEnvDirectory">目录</button>
                </div>
              </div>
              
              <div class="help-box">
                <div><strong>功能介绍：</strong></div>
                <div><strong>说明：</strong>方便快捷的将程序添加到系统变量</div>
                <div><strong>使用：</strong>支持手动输入路径，支持选择程序或者选择目录</div>
                <div><strong>效果：</strong>快速将路径将添加到系统级PATH，重新打开终端即可使用</div>
              </div>
            </div>
          </div>

          <div class="form-group">
            <label>快捷操作</label>
            <button class="win11-button" @click="handleOpenSystemEnvironmentSettings">
              打开系统环境变量设置界面
            </button>
          </div>
        </div>
        <div class="dialog-footer">
          <button class="win11-button" @click="closeEnvDialog">取消</button>
          <button 
            class="win11-button win11-button-primary" 
            @click="handleAddEnvironmentVariable"
            :disabled="!envPath.trim() || isAddingEnv"
          >
            {{ isAddingEnv ? '添加中...' : '添加到PATH' }}
          </button>
        </div>
      </div>
    </div>

    <div class="main-content">
      <div class="content-area">
        <div class="content-header">
          <span class="section-title">服务列表</span>
          <button class="win11-button" @click="loadServices">
            <span class="icon">🔄</span> 刷新
          </button>
        </div>
        
        <div v-if="services.length === 0" class="empty-state">
          <div class="empty-state-icon">⚙️</div>
          <div class="empty-state-text">
            暂无服务<br>
            点击右上角"添加服务"按钮开始创建服务
          </div>
        </div>
        
        <div v-else class="table-container">
          <table class="win11-table">
            <thead>
              <tr>
                <th>服务名称</th>
                <th>状态</th>
                <th>程序路径</th>
                <th>开机自启</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="service in services" :key="service.id" class="table-row">
                <td>
                  <div class="service-name">{{ service.name }}</div>
                  <div class="service-pid">PID: {{ service.pid || 'N/A' }}</div>
                </td>
                <td>
                  <div :class="['service-status', service.status]">
                    <span :class="['status-dot', service.status]"></span>
                    {{ service.status === 'running' ? '运行中' : service.status === 'error' ? '错误' : '已停止' }}
                  </div>
                </td>
                <td>
                  <div class="exe-path">{{ service.exePath }}</div>
                  <div v-if="service.args" class="exe-args">参数: {{ service.args }}</div>
                </td>
                <td>
                  <label class="switch">
                    <input 
                      type="checkbox" 
                      :checked="service.autoStart || false" 
                      @change="handleAutoStartToggle(service.id, $event.target.checked)"
                    >
                    <span class="slider"></span>
                  </label>
                </td>
                <td>
                  <div class="action-buttons">
                    <button 
                      v-if="service.status === 'stopped'"
                      class="win11-button win11-button-small"
                      @click="handleStartService(service.id)"
                      title="启动服务"
                    >
                      ▶️
                    </button>
                    <button 
                      v-else
                      class="win11-button win11-button-small"
                      @click="handleStopService(service.id)"
                      title="停止服务"
                    >
                      ⏹️
                    </button>
                    <button 
                      class="win11-button win11-button-small"
                      @click="handleViewLogs(service.id, service.name)"
                      title="查看日志"
                    >
                      📄
                    </button>
                    <button 
                      class="win11-button win11-button-small win11-delete-button"
                      @click="handleDeleteService(service.id)"
                      title="删除服务"
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
    </div>

    <div class="status-bar">
      <span>
        总计服务: {{ serviceStats.total }} | 
        运行中: {{ serviceStats.running }} | 
        已停止: {{ serviceStats.stopped }}
        {{ adminPrivileges ? ' | 管理员权限' : ' | 普通权限' }}
      </span>
    </div>

    <!-- 删除确认对话框 -->
    <div v-if="isDeleteDialogOpen" class="dialog-overlay" @click.self="isDeleteDialogOpen = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3>确认删除服务</h3>
        </div>
        <div class="dialog-body">
          <p>确定要删除服务 "{{ serviceToDelete?.name }}" 吗？</p>
          <p class="warning-text">服务将被删除！</p>
        </div>
        <div class="dialog-footer">
          <button class="win11-button" @click="isDeleteDialogOpen = false">取消</button>
          <button class="win11-button win11-button-danger" @click="confirmDeleteService">删除</button>
        </div>
      </div>
    </div>

    <!-- 日志查看对话框 -->
    <div v-if="isLogsDialogOpen" class="dialog-overlay" @click.self="isLogsDialogOpen = false">
      <div class="dialog dialog-large">
        <div class="dialog-header">
          <h3>服务日志 - {{ serviceToViewLogs?.name }}</h3>
          <button class="dialog-close" @click="isLogsDialogOpen = false">✕</button>
        </div>
        <div class="dialog-body">
          <div class="log-content">
            {{ serviceLogs || '暂无日志' }}
          </div>
        </div>
        <div class="dialog-footer">
          <button class="win11-button" @click="handleOpenLogsDirectory">📂 打开日志目录</button>
          <button class="win11-button" @click="handleCopyLogsPath">📋 复制路径</button>
          <button class="win11-button win11-button-primary" @click="isLogsDialogOpen = false">关闭</button>
        </div>
      </div>
    </div>

    <!-- Toast 通知 -->
    <div v-if="toast.visible" :class="['toast', toast.type]">
      <div class="toast-title">{{ toast.title }}</div>
      <div v-if="toast.message" class="toast-message">{{ toast.message }}</div>
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
import { EventsOn, EventsOff } from "../wailsjs/runtime/runtime"

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

const toast = ref({
  visible: false,
  title: '',
  message: '',
  type: 'success'
})

let toastTimeout = null

const showToast = (title, message, type = 'success') => {
  if (toastTimeout) clearTimeout(toastTimeout)
  toast.value = { visible: true, title, message, type }
  toastTimeout = setTimeout(() => {
    toast.value.visible = false
  }, 3000)
}

const serviceStats = computed(() => ({
  total: services.value.length,
  running: services.value.filter(s => s.status === 'running').length,
  stopped: services.value.filter(s => s.status === 'stopped').length
}))

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

const handleAppAutoStartToggle = async () => {
  try {
    await SetAutoStart(autoStart.value)
    showToast('成功', `开机自启动已${autoStart.value ? '启用' : '禁用'}`)
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
    newService.value = {
      name: '',
      exePath: '',
      args: '',
      workingDir: ''
    }
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
    
    isEnvDialogOpen.value = false
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

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: 'Segoe UI', -apple-system, BlinkMacSystemFont, sans-serif;
  background-color: #f5f5f5;
  overflow: hidden;
}

#app {
  height: 100vh;
}

.app-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.header {
  display: flex;
  align-items: center;
  padding: 16px 24px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.header-title {
  font-size: 20px;
  font-weight: 600;
  color: #333;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-left: auto;
}

.badge {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.badge-warning {
  background-color: #fff3cd;
  color: #856404;
}

.badge-success {
  background-color: #d4edda;
  color: #155724;
}

.win11-button {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: 1px solid #d1d1d1;
  border-radius: 6px;
  background: white;
  color: #333;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.win11-button:hover {
  background: #f5f5f5;
  border-color: #c1c1c1;
}

.win11-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.win11-button-primary {
  background: #0078d4;
  border-color: #0078d4;
  color: white;
}

.win11-button-primary:hover {
  background: #106ebe;
  border-color: #106ebe;
}

.win11-button-danger {
  background: #d13438;
  border-color: #d13438;
  color: white;
}

.win11-button-danger:hover {
  background: #a80000;
  border-color: #a80000;
}

.win11-button-small {
  padding: 4px 8px;
  font-size: 12px;
}

.win11-delete-button:hover {
  background-color: #fef2f2 !important;
  border-color: #fecaca !important;
  color: #dc2626 !important;
}

.icon {
  font-size: 14px;
}

.main-content {
  flex: 1;
  padding: 24px;
  overflow: hidden;
}

.content-area {
  background: rgba(255, 255, 255, 0.98);
  border-radius: 12px;
  padding: 20px;
  height: 100%;
  overflow: auto;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
}

.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #333;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #666;
}

.empty-state-icon {
  font-size: 64px;
  margin-bottom: 16px;
}

.empty-state-text {
  text-align: center;
  line-height: 1.8;
}

.table-container {
  overflow-x: auto;
}

.win11-table {
  width: 100%;
  border-collapse: collapse;
  border-radius: 8px;
  overflow: hidden;
}

.win11-table thead {
  background: #f3f3f3;
}

.win11-table th {
  padding: 12px 16px;
  text-align: left;
  font-weight: 600;
  color: #333;
  border-bottom: 2px solid #e0e0e0;
}

.win11-table td {
  padding: 12px 16px;
  border-bottom: 1px solid #eee;
  vertical-align: middle;
}

.table-row:hover {
  background-color: #f9f9f9;
}

.service-name {
  font-weight: 600;
  color: #333;
}

.service-pid {
  font-size: 12px;
  color: #666;
}

.service-status {
  display: flex;
  align-items: center;
  gap: 6px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #605e5c;
}

.status-dot.running {
  background-color: #107c10;
}

.status-dot.error {
  background-color: #c42b1c;
}

.exe-path {
  word-break: break-all;
  font-size: 13px;
}

.exe-args {
  font-size: 11px;
  color: #666;
  font-style: italic;
}

.action-buttons {
  display: flex;
  gap: 6px;
}

.switch {
  position: relative;
  display: inline-block;
  width: 40px;
  height: 20px;
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
  border-radius: 20px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 14px;
  width: 14px;
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

.status-bar {
  padding: 10px 24px;
  background: rgba(255, 255, 255, 0.9);
  font-size: 12px;
  color: #666;
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
}

.dialog-large {
  max-width: 800px;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #eee;
}

.dialog-header h3 {
  font-size: 18px;
  font-weight: 600;
  color: #333;
}

.dialog-close {
  background: none;
  border: none;
  font-size: 18px;
  cursor: pointer;
  color: #666;
  padding: 4px 8px;
}

.dialog-close:hover {
  color: #333;
}

.dialog-body {
  padding: 20px;
  overflow-y: auto;
  flex: 1;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 16px 20px;
  border-top: 1px solid #eee;
}

.form-group {
  margin-bottom: 18px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #333;
}

.required {
  color: #d13438;
}

.win11-input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d1d1;
  border-radius: 6px;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
}

.win11-input:focus {
  border-color: #0078d4;
}

.input-with-button {
  display: flex;
  gap: 8px;
}

.input-with-button .win11-input {
  flex: 1;
}

.button-group {
  display: flex;
  gap: 4px;
}

.info-box {
  padding: 12px 16px;
  background: #f3f4f6;
  border-radius: 6px;
  border: 1px solid #e5e7eb;
  font-size: 14px;
  color: #666;
  font-style: italic;
}

.warning-content {
  text-align: center;
}

.warning-text {
  color: #d13438;
  font-weight: 600;
}

.settings-section {
  margin-bottom: 20px;
}

.settings-section > label {
  display: block;
  font-weight: 600;
  color: #333;
  margin-bottom: 10px;
}

.settings-content {
  padding: 12px 16px;
  background: rgba(255, 255, 255, 0.5);
  border-radius: 12px;
  backdrop-filter: blur(10px);
}

.setting-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.setting-row:last-child {
  margin-bottom: 0;
}

.setting-action {
  margin-top: 12px;
}

.app-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.app-title {
  font-weight: 600;
}

.env-info-box {
  padding: 16px;
  background: rgba(255, 255, 255, 0.5);
  border-radius: 12px;
  backdrop-filter: blur(10px);
  border: 1px solid #e5e7eb;
}

.help-box {
  margin-top: 12px;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 6px;
  border: 1px solid #e9ecef;
  font-size: 12px;
  color: #666;
  line-height: 1.8;
}

.log-content {
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
}

.toast {
  position: fixed;
  bottom: 80px;
  right: 24px;
  padding: 16px 20px;
  border-radius: 8px;
  background: white;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);
  z-index: 2000;
  animation: slideIn 0.3s ease;
}

.toast.success {
  border-left: 4px solid #107c10;
}

.toast.error {
  border-left: 4px solid #d13438;
}

.toast-title {
  font-weight: 600;
  color: #333;
}

.toast-message {
  margin-top: 4px;
  font-size: 13px;
  color: #666;
}

@keyframes slideIn {
  from {
    transform: translateX(100%);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

.slide-in {
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
