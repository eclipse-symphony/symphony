using System.Collections.ObjectModel;
using System.ComponentModel;
using Windows.UI.Xaml.Controls;
using Windows.UI.Xaml;
using System.Linq;
using System;
using System.Threading.Tasks;
using System.Net.Http;
using System.Text;
using Newtonsoft.Json;

namespace ContosoCafeteriaKiosk
{
    public sealed partial class MainPage : Page, INotifyPropertyChanged
    {
        public ObservableCollection<MenuItem> MenuItems { get; set; }
        public ObservableCollection<OrderItem> OrderItems { get; set; }

        private double _totalPrice;
        public double TotalPrice
        {
            get => _totalPrice;
            set
            {
                if (_totalPrice != value)
                {
                    _totalPrice = value;
                    OnPropertyChanged(nameof(TotalPrice));
                    OnPropertyChanged(nameof(TotalPriceDisplay));
                }
            }
        }

        public string TotalPriceDisplay => TotalPrice.ToString("C2");

        public MainPage()
        {
            this.InitializeComponent();
            OrderItems = new ObservableCollection< OrderItem>();
            MenuItems = new ObservableCollection<MenuItem>()
            {
                new MenuItem(){ Name="FruitSmoothie", Price=2.99, ImagePath="Assets/FruitSmoothie.png" },
                new MenuItem(){ Name="Soda", Price=0.99, ImagePath="Assets/Soda.png" },
                new MenuItem(){ Name="IceCream", Price=1.49, ImagePath="Assets/IceCream.png" },
                new MenuItem(){ Name="Cookie", Price=1.49, ImagePath="Assets/Cookie.png" },
                new MenuItem(){ Name="PizzaSlice", Price=1.99, ImagePath="Assets/PizzaSlice.png" },
                new MenuItem(){ Name="Hotdog", Price=1.99, ImagePath="Assets/Hotdog.png" },
                new MenuItem(){ Name="Nachos", Price=2.99, ImagePath="Assets/Nachos.png" },
            };
            this.DataContext = this;
        }

        private void GridView_ItemClick(object sender, ItemClickEventArgs e)
        {
            var item = e.ClickedItem as MenuItem;
            if (item != null)
            {
                TotalPrice += item.Price;
                var orderItem = OrderItems.FirstOrDefault(o => o.Name == item.Name);
                if (orderItem == null)
                {
                    OrderItems.Add(new OrderItem { Name = item.Name, Price = item.Price, Quantity = 1 });
                }
                else
                {
                    orderItem.Quantity++;
                }
            }
        }
        private async Task SendOrderAsync(object order)
        {
            using (HttpClient client = new HttpClient())
            {
                client.BaseAddress = new Uri("http://localhost:5000");
                client.DefaultRequestHeaders.Accept.Clear();
                client.DefaultRequestHeaders.Accept.Add(new System.Net.Http.Headers.MediaTypeWithQualityHeaderValue("application/json"));

                string jsonOrder = JsonConvert.SerializeObject(order);
                StringContent content = new StringContent(jsonOrder, Encoding.UTF8, "application/json");

                try
                {
                    HttpResponseMessage response = await client.PostAsync("http://localhost:5000/orders", content);
                    response.EnsureSuccessStatusCode();
                    System.Diagnostics.Debug.WriteLine("Order sent successfully.");
                }
                catch (HttpRequestException ex)
                {
                    System.Diagnostics.Debug.WriteLine($"Request error: {ex.Message}");
                }
                catch (Exception ex)
                {
                    System.Diagnostics.Debug.WriteLine($"An error occurred: {ex.Message}");
                }
                finally
                {
                    System.Diagnostics.Debug.WriteLine("SendOrderAsync method execution completed.");
                }
            }
        }

        private async void PlaceOrderButton_Click(object sender, RoutedEventArgs e)
        {
            string orderDetails = "Your order:\n";
            foreach (var orderItem in OrderItems)
            {
                orderDetails += $"{orderItem.Name} x{orderItem.Quantity} - {orderItem.PriceDisplay}\n";
            }

            orderDetails += $"\nTotal Price: {TotalPriceDisplay}";

            ContentDialog dialog = new ContentDialog()
            {
                Title = "Order Placed",
                Content = orderDetails,
                CloseButtonText = "OK"
            };
            _ = dialog.ShowAsync();

            var orderItemsList = OrderItems.Select(item => new OrderItemDto
            {
                name = item.Name,
                quantity = item.Quantity,
            }).ToList();

            var order = new
            {
                customerName = "test",
                items = orderItemsList,
                time = DateTime.Now.ToString("yyyy-MM-dd HH:mm:ss")
            };

            await SendOrderAsync(order);

            OrderItems.Clear();
            TotalPrice = 0.00;
        }

        public event PropertyChangedEventHandler PropertyChanged;
        private void OnPropertyChanged(string propertyName)
        {
            PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(propertyName));
        }
    }

    public class OrderItemDto
    {
        public string name { get; set; }
        public int quantity { get; set; }
    }

    public class MenuItem
    {
        public string Name { get; set; }
        public double Price { get; set; }
        public string ImagePath { get; set; }

        public string PriceDisplay => Price.ToString("C2");

    }
    public class OrderItem: INotifyPropertyChanged
    {
        public string Name { get; set; }
        public double Price { get; set; }
        private int _quantity;
        public int Quantity
        {
            get => _quantity;
            set
            {
                if (_quantity != value)
                {
                    _quantity = value;
                    OnPropertyChanged(nameof(Quantity));
                }
            }
        }

        public string PriceDisplay => Price.ToString("C2");

        public event PropertyChangedEventHandler PropertyChanged;
        private void OnPropertyChanged(string propertyName)
        {
            PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(propertyName));
        }
    }
}