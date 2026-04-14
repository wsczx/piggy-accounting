import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import {
  GetByDateRange as GetRecords, SearchRecords as SearchRecordsAPI,
  Add as AddRecord, AddWithAccount as AddRecordWithAccount,
  Delete as DeleteRecord, Update as UpdateRecord, GetByID as GetRecord,
  GetMonthlyStats, GetMonthlyCategoryStats, GetDailyStats,
  GetYearlyStats, GetYearlyCategoryStats, GetMonthlyTrend
} from '../../wailsjs/go/service/RecordService'
import {
  GetAll as GetCategories, Add as AddCategory, Delete as DeleteCategory
} from '../../wailsjs/go/service/CategoryService'
import {
  SetBudget, GetBudgetInfo, DeleteBudget
} from '../../wailsjs/go/service/BudgetService'
import { GetCurrentMonth, GetCurrentDate } from '../../wailsjs/go/main/App'
import { createLogger } from '../utils/logger'

const log = createLogger('Store')

export interface TagInfo {
  id: number
  name: string
  color: string
}

export interface Record {
  id: number
  type: string
  amount: number
  category: string
  note: string
  date: string
  created_at: string
  account_id?: number
  tags?: TagInfo[]
}

export interface Category {
  id: number
  name: string
  icon: string
  type: string
  is_system: boolean
}

export interface MonthlyStats {
  total_income: number
  total_expense: number
  balance: number
  month: string
}

export interface CategoryStat {
  category: string
  category_icon: string
  amount: number
  percentage: number
}

export interface DailyStat {
  date: string
  total_income: number
  total_expense: number
}

export interface MonthlyTrend {
  month: string
  total_income: number
  total_expense: number
}

export interface YearlyStats {
  year: string
  total_income: number
  total_expense: number
  balance: number
}

export interface SearchParams {
  keyword?: string
  type?: string
  category?: string
  category_ids?: string
  tag_id?: number
}

export interface SearchResult {
  records: Record[]
  total: number
  page: number
  limit: number
}

export interface BudgetInfo {
  budget_type: string
  year: number
  month: number
  budget_amount: number
  spent: number
  remaining: number
  percentage: number
}

export const useAccountingStore = defineStore('accounting', () => {
  // State
  const records = ref<Record[]>([])
  /** 全量记录缓存（用于胶囊筛选等场景，不受搜索/分页影响） */
  const allRecordsCache = ref<Record[]>([])
  const categories = ref<Category[]>([])
  const monthlyStats = ref<MonthlyStats>({ total_income: 0, total_expense: 0, balance: 0, month: '' })
  const categoryStats = ref<CategoryStat[]>([])
  const dailyStats = ref<DailyStat[]>([])
  const currentMonth = ref('')
  const loading = ref(false)
  const error = ref<string | null>(null)
  const initialized = ref(false)  // 标记是否已初始化

  // 搜索/分页状态
  const searchPage = ref(1)
  const searchTotal = ref(0)
  const searchHasMore = ref(false)
  const searchLoading = ref(false)
  const pageSize = 20

  // 预算状态
  const monthlyBudget = ref<BudgetInfo | null>(null)
  const yearlyBudget = ref<BudgetInfo | null>(null)

  // 年度统计状态
  const yearlyStats = ref<YearlyStats>({ year: '', total_income: 0, total_expense: 0, balance: 0 })
  const yearlyCategoryStats = ref<CategoryStat[]>([])
  const monthlyTrend = ref<MonthlyTrend[]>([])

  // Computed
  const expenseCategories = computed(() => categories.value.filter(c => c.type === 'expense'))
  const incomeCategories = computed(() => categories.value.filter(c => c.type === 'income'))
  const monthName = computed(() => {
    if (!currentMonth.value) return ''
    const [y, m] = currentMonth.value.split('-')
    return `${y}年${parseInt(m)}月`
  })

  // 记录相关 - 使用搜索分页 API
  /** 初始化全量记录缓存（用于胶囊分类显示，不受日期范围限制） */
  async function loadRecords(startDate = '', endDate = '', reset = true) {
    try {
      searchLoading.value = true
      if (reset) {
        searchPage.value = 1
        records.value = []
      }
      log.info('加载记录', 'page:', searchPage.value, startDate, '~', endDate)
      const result = await SearchRecordsAPI({
        start_date: startDate,
        end_date: endDate,
        type: '',
        category: '',
        category_ids: '',
        keyword: '',
        account_id: 0,
        tag_id: 0,
        page: searchPage.value,
        limit: pageSize,
      })
      const newRecords = result?.records ?? []
      if (reset) {
        records.value = newRecords
      } else {
        // 追加模式：去重（避免刷新后重复）
        const existingIds = new Set(records.value.map(r => r.id))
        const unique = newRecords.filter(r => !existingIds.has(r.id))
        records.value = [...records.value, ...unique]
      }
      searchTotal.value = result?.total ?? 0
      searchHasMore.value = records.value.length < searchTotal.value
      log.info('记录加载完成, 当前:', records.value.length, '总计:', searchTotal.value)
    } catch (e) {
      log.error('加载记录失败', e)
      error.value = `加载记录失败: ${e}`
    } finally {
      searchLoading.value = false
    }
  }

  async function loadMoreRecords() {
    if (searchLoading.value || !searchHasMore.value) return
    searchPage.value++
    const month = currentMonth.value
    await loadRecords(month + '-01', month + '-31', false)
  }

  async function searchRecords(params: SearchParams, reset = true) {
    try {
      searchLoading.value = true
      if (reset) {
        searchPage.value = 1
        records.value = []
      }
      const month = currentMonth.value
      log.info('搜索记录', params)
      const result = await SearchRecordsAPI({
        start_date: month + '-01',
        end_date: month + '-31',
        type: params.type || '',
        category: params.category || '',
        category_ids: params.category_ids || '',
        keyword: params.keyword || '',
        account_id: 0,
        tag_id: params.tag_id || 0,
        page: searchPage.value,
        limit: pageSize,
      })
      const newRecords = result?.records ?? []
      if (reset) {
        records.value = newRecords
      } else {
        const existingIds = new Set(records.value.map(r => r.id))
        const unique = newRecords.filter(r => !existingIds.has(r.id))
        records.value = [...records.value, ...unique]
      }
      searchTotal.value = result?.total ?? 0
      searchHasMore.value = records.value.length < searchTotal.value
      log.info('搜索完成, 当前:', records.value.length, '总计:', searchTotal.value)
    } catch (e) {
      log.error('搜索记录失败', e)
      error.value = `搜索失败: ${e}`
    } finally {
      searchLoading.value = false
    }
  }

  async function addRecord(type: string, category: string, note: string, date: string, amount: number, accountId: number = 0): Promise<number> {
    try {
      log.info('添加记录', type, category, amount, 'accountId', accountId)
      let recordId: number
      if (accountId > 0) {
        recordId = await AddRecordWithAccount(type, category, note, date, amount, accountId)
      } else {
        recordId = await AddRecord(type, category, note, date, amount)
      }
      log.info('添加记录成功, id:', recordId)
      await refreshAll()
      return recordId
    } catch (e) {
      log.error('添加记录失败', e)
      throw e // 重新抛出让调用者处理
    }
  }

  async function deleteRecord(id: number) {
    try {
      await DeleteRecord(id)
      await refreshAll()
    } catch (e) {
      log.error('删除记录失败', e)
      throw e
    }
  }

  async function updateRecord(id: number, type: string, category: string, note: string, date: string, amount: number) {
    try {
      await UpdateRecord(id, type, category, note, date, amount)
      await refreshAll()
    } catch (e) {
      log.error('更新记录失败', e)
      throw e
    }
  }

  // 类别相关
  async function loadCategories() {
    try {
      log.info('加载类别(全部)')
      const result = await GetCategories()
      categories.value = result ?? []
      log.info('类别加载完成, 数量:', categories.value.length,
        '支出:', categories.value.filter(c => c.type === 'expense').length,
        '收入:', categories.value.filter(c => c.type === 'income').length)
    } catch (e) {
      log.error('加载类别失败', e)
      error.value = `加载类别失败: ${e}`
    }
  }

  async function addCategory(name: string, icon: string, type: string) {
    try {
      await AddCategory(name, icon, type)
      await loadCategories()
    } catch (e) {
      log.error('添加类别失败', e)
      throw e
    }
  }

  async function deleteCategory(id: number) {
    try {
      await DeleteCategory(id)
      await loadCategories()
    } catch (e) {
      log.error('删除类别失败', e)
      throw e
    }
  }

  // 统计相关
  async function loadMonthlyStats(month: string) {
    try {
      monthlyStats.value = await GetMonthlyStats(month)
    } catch (e) {
      log.error('加载月度统计失败', e)
    }
  }

  async function loadCategoryStats(month: string, type: string) {
    try {
      categoryStats.value = (await GetMonthlyCategoryStats(month, type)) ?? []
    } catch (e) {
      log.error('加载类别统计失败', e)
    }
  }

  async function loadDailyStats(month: string) {
    try {
      dailyStats.value = (await GetDailyStats(month)) ?? []
    } catch (e) {
      log.error('加载每日统计失败', e)
    }
  }

  // 年度统计
  async function loadYearlyStats(year: string) {
    try {
      yearlyStats.value = await GetYearlyStats(year)
    } catch (e) {
      log.error('加载年度统计失败', e)
    }
  }

  async function loadYearlyCategoryStats(year: string, type: string) {
    try {
      yearlyCategoryStats.value = (await GetYearlyCategoryStats(year, type)) ?? []
    } catch (e) {
      log.error('加载年度类别统计失败', e)
    }
  }

  async function loadMonthlyTrend(year: string) {
    try {
      monthlyTrend.value = (await GetMonthlyTrend(year)) ?? []
    } catch (e) {
      log.error('加载月度趋势失败', e)
    }
  }

  // 预算相关
  async function loadBudgetInfo() {
    try {
      if (!currentMonth.value) return
      const [y, m] = currentMonth.value.split('-').map(Number)
      const monthNum = m as number

      // 并行加载月度和年度预算
      const [monthlyResult, yearlyResult] = await Promise.all([
        GetBudgetInfo('monthly', y, monthNum),
        GetBudgetInfo('yearly', y, 0),
      ])

      monthlyBudget.value = monthlyResult ?? null
      yearlyBudget.value = yearlyResult ?? null
      log.info('预算信息已加载')
    } catch (e) {
      log.error('加载预算信息失败', e)
    }
  }

  async function setBudget(budgetType: string, year: number, month: number, amount: number) {
    try {
      log.info('设置预算', budgetType, year, month, amount)
      await SetBudget(budgetType, year, month, amount)
      await loadBudgetInfo()
      log.info('预算设置成功')
    } catch (e) {
      log.error('设置预算失败', e)
      throw e
    }
  }

  async function deleteBudget(budgetType: string, year: number, month: number) {
    try {
      log.info('删除预算', budgetType, year, month)
      await DeleteBudget(budgetType, year, month)
      await loadBudgetInfo()
      log.info('预算删除成功')
    } catch (e) {
      log.error('删除预算失败', e)
      throw e
    }
  }

  // 初始化和刷新
  async function init() {
    // 如果已经初始化过，直接刷新数据即可
    if (initialized.value && !loading.value) {
      log.info('Store 已初始化，执行刷新')
      await refreshAll()
      return
    }
    // 如果正在初始化中，等待完成
    if (loading.value) {
      log.info('Store 正在初始化中，跳过')
      return
    }
    loading.value = true
    error.value = null
    try {
      log.info('Store 初始化开始')
      currentMonth.value = await GetCurrentMonth()
      log.info('当前月份:', currentMonth.value)

      // 先加载类别（确保 categories 就绪，避免 RecordPanel 等待）
      await loadCategories()
      log.info('类别已加载, 开始加载记录和统计')

      await refreshAll()
      initialized.value = true
      log.info('Store 初始化完成')
    } catch (e) {
      log.error('Store 初始化失败', e)
      error.value = `初始化失败: ${e}`
    } finally {
      loading.value = false
    }
  }

  async function refreshAll() {
    try {
      allRecordsCache.value = [] // 清除全量缓存，确保数据一致性
      if (!currentMonth.value) currentMonth.value = await GetCurrentMonth()
      const month = currentMonth.value
      const startDate = month + '-01'
      const endDate = month + '-31'

      // 统计数据并行请求
      const [, statsResult, dailyResult, categoryResult] = await Promise.all([
        loadRecords(startDate, endDate, true),
        GetMonthlyStats(month),
        GetDailyStats(month),
        GetMonthlyCategoryStats(month, 'expense'),
      ])

      monthlyStats.value = statsResult ?? { total_income: 0, total_expense: 0, balance: 0, month }
      dailyStats.value = dailyResult ?? []
      categoryStats.value = categoryResult ?? []

      // 预算信息（不阻塞主流程）
      loadBudgetInfo()

      // 单独加载全量记录缓存（用于胶囊筛选等场景，不受分页和筛选影响）
      try {
        const cacheResult = await SearchRecordsAPI({
          start_date: startDate,
          end_date: endDate,
          type: '', category: '', category_ids: '', keyword: '',
          account_id: 0, tag_id: 0,
          page: 1, limit: 9999,
        })
        allRecordsCache.value = cacheResult?.records ?? []
        log.info('全量记录缓存更新完成, 共', allRecordsCache.value.length, '条')
      } catch (e) {
        log.error('加载全量记录缓存失败', e)
      }
    } catch (e) {
      log.error('刷新数据失败', e)
    }
  }

  /** 填充全量记录缓存（供外部页面调用，如 RecordsView） */
  function fillAllRecordsCache(data: Record[]) {
    allRecordsCache.value = data
  }

  function changeMonth(month: string) {
    log.info('切换月份:', month)
    currentMonth.value = month
    refreshAll()
  }

  return {
    records, allRecordsCache, categories, monthlyStats, categoryStats, dailyStats,
    currentMonth, loading, error, initialized, monthName,
    searchPage, searchTotal, searchHasMore, searchLoading,
    monthlyBudget, yearlyBudget,
    yearlyStats, yearlyCategoryStats, monthlyTrend,
    expenseCategories, incomeCategories,
    loadRecords, loadMoreRecords, searchRecords,
    addRecord, deleteRecord, updateRecord,
    loadCategories, addCategory, deleteCategory,
    loadMonthlyStats, loadCategoryStats, loadDailyStats,
    loadYearlyStats, loadYearlyCategoryStats, loadMonthlyTrend,
    loadBudgetInfo, setBudget, deleteBudget,
    init, refreshAll, changeMonth, fillAllRecordsCache
  }
})
