<script lang="ts" setup>
import Highcharts from 'highcharts'
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { DateTime } from 'luxon'

const chartContainer = ref<HTMLElement | null>(null)
let chart: Highcharts.Chart | undefined
const isLoading = ref<boolean>(true)
const chartData = ref<number[]>([])
const graphCategories = ref<string[]>([])
const selectedView = ref<string>('week')
const selectedDuration = ref<string>('1 months')

async function getChartData({
  selectedView,
  selectedDuration,
}: {
  selectedView: string
  selectedDuration: string
}) {
  try {
    isLoading.value = true
    const baseURL = `${import.meta.env.VITE_SERVER_BASE_URL}/completed-task-count`
    const searchParams = new URLSearchParams({ v: selectedView, d: selectedDuration })
    const url = `${baseURL}?${searchParams.toString()}`
    const response = await fetch(url)
    if (response.status !== 200) {
      throw new Error("Something wen't wrong")
    }

    const { completedCount } = await response.json()
    const categories: string[] = []
    const seriesData: number[] = []
    for (const item of completedCount) {
      const { completedAt, taskCount } = item
      categories.push(DateTime.fromISO(completedAt).toFormat('dd-LLL-yyyy'))
      seriesData.push(taskCount)
    }

    graphCategories.value = categories
    chartData.value = seriesData
  } catch (e) {
    console.error(e)
  } finally {
    isLoading.value = false
  }
}

function loadChartData() {
  if (chartContainer.value && chartData.value.length > 0) {
    Highcharts.setOptions({
      lang: {
        decimalPoint: '.',
        thousandsSep: ',',
        months: [
          'January',
          'February',
          'March',
          'April',
          'May',
          'June',
          'July',
          'August',
          'September',
          'October',
          'November',
          'December',
        ],
        weekdays: ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'],
        shortMonths: [
          'Jan',
          'Feb',
          'Mar',
          'Apr',
          'May',
          'Jun',
          'Jul',
          'Aug',
          'Sep',
          'Oct',
          'Nov',
          'Dec',
        ],
      },
    })

    chart = Highcharts.chart(chartContainer.value, {
      chart: {
        type: 'line',
      },
      title: {
        text: 'Task Completion Summary',
        align: 'center',
      },
      yAxis: {
        title: {
          text: 'Total completed tasks',
        },
      },
      xAxis: {
        categories: graphCategories.value,
      },
      legend: {
        layout: 'horizontal',
        align: 'center',
        verticalAlign: 'bottom',
      },
      plotOptions: {
        series: {
          label: {
            connectorAllowed: false,
          },
        },
      },
      credits: {
        enabled: false,
      },
      series: [
        {
          name: 'Task completion count',
          type: 'line',
          color: 'oklch(54.6% 0.245 262.881)',
          data: chartData.value,
        },
      ],
      responsive: {
        rules: [
          {
            condition: {
              maxWidth: 500,
            },
            chartOptions: {
              legend: {
                layout: 'horizontal',
                align: 'center',
                verticalAlign: 'bottom',
              },
            },
          },
        ],
      },
    })
  }
}

onMounted(() => {
  getChartData({ selectedDuration: selectedDuration.value, selectedView: selectedView.value }).then(
    () => {
      loadChartData()
    },
  )
})

onUnmounted(() => {
  chart?.destroy()
})

watch([selectedView, selectedDuration], ([newSelectedView, newSelectedDuration]) => {
  getChartData({ selectedDuration: newSelectedDuration, selectedView: newSelectedView })
    .then(() => {
      loadChartData()
    })
})
</script>

<template>
  <figure class="chart-wrapper">
    <div>
      <select name="view" id="view" class="view-select" v-model="selectedView">
        <option value="week">Weekly</option>
        <option value="month">Monthly</option>
      </select>

      <select name="duration" id="duration" class="duration-select" v-model="selectedDuration">
        <option value="1 months">This month</option>
        <option value="2 months">2 months</option>
        <option value="3 months">3 months</option>
      </select>
    </div>
    <div ref="chartContainer" class="chart-container"></div>
  </figure>
</template>

<style scoped>
.chart-wrapper {
  width: 100%;
  box-shadow:
    0 4px 6px -1px rgb(0 0 0 / 0.1),
    0 2px 4px -2px rgb(0 0 0 / 0.1);
  border-radius: 0.25rem;
  overflow: hidden;
  background-color: #fff;
  padding: 1rem;
}

.chart-container {
  width: 100%;
  height: 400px;
  margin: 0 auto;
}

.view-select,
.duration-select {
  outline: none;
  border: 2px solid var(--primary);
  border-radius: 0.375rem;
  padding: 0.5rem 1rem;
  margin-right: 1rem;
  background-color: #f9fafb;
  color: #374151;
  font-size: 1rem;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.view-select:focus,
.duration-select:focus {
  border-color: var(--primary);
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.2);
}

@media (max-width: 600px) {
  .view-select,
  .duration-select {
    width: 100%;
    margin-bottom: 0.5rem;
    margin-right: 0;
  }
}
</style>
