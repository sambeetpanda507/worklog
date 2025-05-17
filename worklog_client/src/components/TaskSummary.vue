<script setup lang="ts">
import type { ITaskStatusSummary } from '@/interfaces'
import Highcharts from 'highcharts'
import { onMounted, onUnmounted, ref } from 'vue'

type ChartData = {
  name: string
  y: number
}

const chartContainer = ref<HTMLElement | null>(null)
let chart: Highcharts.Chart | undefined
const isLoading = ref<boolean>(true)
const chartData = ref<ChartData[]>([] as ChartData[])

const getChartData = async () => {
  try {
    isLoading.value = true
    const response = await fetch('http://localhost:5001/status-summary', {
      method: 'GET',
      headers: { Accept: 'application/json' },
    })

    if (response.status != 200) {
      console.log('Some error occurred while fetching task status summary data')
    } else {
      const { statusSummary }: { statusSummary: ITaskStatusSummary[] } = await response.json()
      chartData.value = statusSummary.map(({ taskStatus, percentage }) => ({
        name: taskStatus.toUpperCase(),
        y: percentage,
      }))
    }
  } catch (e) {
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  getChartData().then(() => {
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
          type: 'pie',
        },
        title: {
          text: 'Task Status Summary',
        },
        tooltip: {
          pointFormat: '{series.name}: <b>{point.percentage:.0f}' + '%' + '</b>',
        },
        accessibility: {
          point: {
            valueSuffix: '%',
          },
        },
        plotOptions: {
          pie: {
            allowPointSelect: false,
            depth: 35,
            borderRadius: 10,
            innerSize: '65%',
            dataLabels: [
              {
                enabled: true,
                distance: 20,
                format: '{point.name}',
              },
              {
                enabled: true,
                distance: -30,
                format: '{point.percentage:.0f}%',
                style: {
                  fontSize: '0.8em',
                },
              },
            ],
          },
        },
        credits: {
          enabled: false,
        },
        series: [
          {
            name: 'STATUS',
            data: chartData.value,
            type: 'pie',
          },
        ],
        responsive: {
          rules: [
            {
              condition: {
                maxWidth: 300,
              },
              chartOptions: {
                chart: {
                  height: 300,
                },
              },
            },
          ],
        },
      })
    }
  })
})

onUnmounted(() => {
  chart?.destroy()
})
</script>

<template>
  <figure class="chart-wrapper">
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
}

.chart-container {
  width: 100%;
  height: 400px;
  margin: 0 auto;
}
</style>
