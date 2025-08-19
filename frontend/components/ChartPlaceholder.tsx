/**
 * This is a placeholder for future chart components using react-chartjs-2.
 */
import { Chart as ChartJS, CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend } from 'chart.js';
import { Bar } from 'react-chartjs-2';

ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend);

export default function ChartPlaceholder() {
  const data = {
    labels: ['Jan', 'Feb', 'Mar', 'Apr'],
    datasets: [
      {
        label: 'Demo Data',
        data: [12, 19, 3, 5],
        backgroundColor: '#6366f1',
      },
    ],
  };
  const options = {
    responsive: true,
    plugins: {
      legend: { display: false },
      title: { display: true, text: 'Chart.js Demo' },
    },
  };
  return (
    <div className="bg-white rounded-xl p-6 shadow mb-6">
      <Bar data={data} options={options} />
    </div>
  );
}