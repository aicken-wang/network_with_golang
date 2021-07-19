# `Network Programming with Go`

# 第一部分



# 第二部分

## **套接字高级编程**

### 第三章

3.1.1 **可靠的** **TCP** **数据流**

3.1.2 **什么使** **TCP** **可靠？**

3.1.3 **使用** **TCP** **会话**

## 通过使用Go的标准库建立TCP连接

​	Go的标准库中的网络软件包包括对创建基于TCP的服务器和能够连接到这些服务器的客户端的良好支持。即便如此，您还是有责任确保您正确地处理连接。您的软件应该注意传入的数据，并始终努力优雅地关闭连接。让我们编写一个TCP服务器，它可以监听传入的TCP连接、从客户端启动连接、接受和异步处理每个连接、交换数据并终止连接

- 绑定、监听和接受连接（**Binding, Listening for, and Accepting Connections**）

要创建能够监听传入连接（称为listener）的TCP服务器，请使用`net.Listen`功能。此函数将返回一个实现网络的对象。监听器接口。清单3-1显示了监听器的创建情况。

```go
//清单3-1 使用一个由操作系统随机分配的端口，在127.0.0.1上创建一个监听器
package ch03

import (
 "net"
 "testing"
)
func TestListener(t *testing.T) {
 // 1、net.Listen的返回值listener，当net.Listen 接口监听成功会返回listener
 // 2、net.Listen的第一个参数:tcp 协议类型
 // 3、net.Listen的第二个参数:ip地址和端口
 listener, err := net.Listen("tcp", "127.0.0.1:0")
 	if err != nil {
 	t.Fatal(err)
 }
 // 4 defer延迟是否socket资源
 defer func() { _ = listener.Close() }()
 // 5 打印server监听地址
 t.Logf("bound to %q", listener.Addr())
}

```

​	该`net.Listen`函数接受网络类型2和由冒号3分隔的IP地址和端口。该函数返回一个`net.Listen`接口1和一个错误接口。如果函数成功返回，则监听器将绑定到指定的IP地址和端口。绑定意味着操作系统已将给定IP地址上的端口独家分配给监听器。该操作系统不允许其他进程监听绑定端口上的传入流量。如果尝试将监听器绑定到当前绑定的端口，则为net。监听器将返回一个错误。

​	可以选择保留IP地址和端口参数为空。如果端口为零或空，则Go将随机分配一个端口号。您可以通过调用其Addr方法5来检索监听器的地址。同样地，如果您省略了IP地址，则监听器将被绑定到系统上的所有单播和任何单播IP地址。省略IP地址和端口，或者将第二个参数的冒号传递给net。监听，将使您的监听器使用随机端口绑定到所有单播和任何单播IP地址。

​	在大多数情况下，您应该使用tcp作为网络的网络类型。监听者的第一个参数。您可以通过传入tcp4来将监听器限制为IPv4地址，或者通过传入tcp6来独家绑定到IPv6地址。你应该始终细致彻底的通过调用它的Close方法4来优雅地关闭监听器，如果它对你的代码有意义的话，通常是延迟的。当然，这是一个测试用例，当测试完成时，Go会摧毁监听器，但这仍然是一个很好的实践。未能关闭监听器可能会导致代码中的内存泄漏或死锁，因为对监听器的Acceet方法的调用可能会无限期阻塞。关闭监听器将立即取消阻止对Accept方法的调用。

```go
/*
	清单3-2演示了监听器如何接受传入的TCP连接
*/ 
// 1 使用 for-loop来处理新的连接
for { 
    // 2 accpet 有新连接不会阻塞当前for-loop
	conn, err := listener.Accept() //3 accept 接收新的连接，没有新的client连接会阻塞
 	if err != nil {
 		return err
 	}
    // 4 开启一个goroutine
	go func(c net.Conn) {
     // 5 使用defer延迟释放socket资源
 		defer c.Close()
 	  // 您的代码将处理这里的连接
 	}(conn)
 }
```



​	除非您只接受一个传入连接，否则您需要使用循环1，这样您的服务器将接收每个传入连接，在程序中处理它，然后重新循环，准备接受下一个连接。串行接受连接是完全可以接受的，而且，但在此之外，您应该使用`goroutine`来处理每个连接。如果您的用例需要，您当然可以在接受连接后编写序列化代码，但它将非常低效，并且无法利用Go的优势。我们通过调用监听器的“Accept”方法2来启动for循环。此方法将阻止，直到监听器检测到传入连接并完成客户端与服务器之间的TCP握手过程。该调用返回一个`net.TCPConn`接口3 `conn`和一个err。例如，如果握手失败或监听器关闭，则错误接口将为非零( err ! = nil )

​	连接接口的底层类型是指向网络的指针`net.TCPConn`，因为您正在接受TCP连接。连接接口表示`TCP`连接的服务器端。在大多数情况下， `net.Conn`提供了与客户端进行一般交互所需的所有方法。然而，这个`net.TCPConn`对象提供了我们将在第4章中介绍的附加功能

​	要同时处理客户端连接，请剥离一个基本程序来异步处理每个连接4，以便监听器可以为下一个传入的连接做好准备。然后，在程序退出之前调用连接的关闭方法5，通过向服务器发送`FIN`包优雅地终止连接。

## 正在与服务器建立连接

从客户端方面来看，Go的标准库`net`包使接触并与服务器建立连接成为一个简单的事情。清单3-3是一个测试，它演示了使用随机端口上的服务器监听127.0.0.1启动TCP连接的过程

```go
// 清单3-3：建立与127.0.0.1的连接
package ch03

import (
	"io"
	"net"
	"testing"
)

func TestDial(t *testing.T) {
	// 在一个随机端口上创建一个监听器
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	// 1 启动一个goroutine来处理 Reatcor模式
	go func() {
		defer func() {
			done <- struct{}{}
		}()

		for {
			//	2 监听新的客户端
			conn, err := listener.Accept()
			if err != nil {
				t.Log(err)
				return
			}

			// 3 异步读取数据
			go func(c net.Conn) {
				defer func() {
					c.Close()
					done <- struct{}{}
				}()

				buf := make([]byte, 1024)
				for {
					// 4 读取socket缓冲区中的数据
					n, err := c.Read(buf)
					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}
						return
					}
					t.Logf("received: %q", buf[:n])
				}
			}(conn)
		}
	}()
	// 5 conn客户端连接对象
	// 6 net.Dial第一个参数
	// 7 net.Dial第二个参数
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	// 8 关闭连接
	conn.Close()
	<-done
	// 9 关闭监听
	listener.Close()
	<-done
}
```

​	首先在IP地址127.0.0.1上创建监听器，客户端将连接到。您完全省略了端口号，因此Go将随机选择一个可用的端口。然后，在基本程序1中分离监听器，以便在测试后期使用客户端的连接。监听器的程序程序包含像清单3-2这样的代码，用于在循环中接收传入的TCP连接，将每个连接旋转到自己的程序中。(我们经常调用这个`goroutine`处理程序。我将很快解释处理程序的实现细节，但它一次将从套接字读取1024字节并记录收到的内容。）

标准库得`net.Dial`函数就像`net.Listen`的功能，它接受tcp等网络6、IP地址和端口组合7在本例中，是它尝试连接到`listener`的IP地址和端口。您可以使用主机名来代替IP地址和服务名称，如http，来代替端口号。如果主机名解析为多个IP地址，则Go将尝试连接到每个IP地址，直到连接成功或所有IP地址都已用尽为止。由于IPv6地址包含冒号分隔符，IPv6地址必须用方括号括起来。例如，“[2001:ed27::1]:https”在IPv6地址2001:ed27::1处指定了端口443。`net.Dial` (拨号)返回一个连接对象5和一个错误信息。

​	现在您已成功建立了与监听器的连接，您可以从客户端开始优雅地终止连接8。在接收到FIN数据包之后，读取方法4返回`io.EOF`(Linux一切皆文件，socket文件描述符读到文件尾部)错误，向监听器的代码指示您关闭了连接的一侧。连接的处理程序3退出，在退出时调用连接的关闭方法。这将向连接发送一个FIN数据包，从而完成TCP会话的正常终止

​	最后，请关闭监听器(第9处)。第2处的监听器的Accept方法立即解除阻塞并返回错误。这个错误并不一定是失败，所以您只需记录它并继续前进。它不会导致您的测试失败。监听器的goroutine退出，测试完成.

### 理解超时和临时错误

```
In a perfect world, your connection attempts will immediately succeed, and all read and write attempts will never fail. But you need to hope for the best and prepare for the worst. You need a way to determine whether an error is temporary or something that warrants termination of the connection altogether. The error interface doesn't provide enough information to make that determination. Thankfully, Go's net package provides more insight if you know how to use it
```

在一个完美的世界中，您的连接尝试将立即成功，而所有的读写尝试将永远不会失败。但你需要抱最好的希望，做最坏的打算。您需要一种方法来确定错误是临时的还是足以终止整个连接的。错误接口没有提供足够的信息来进行判断。值得庆幸的是，如果您知道如何使用，那么Go的网络包提供了更多的见解

```
Errors returned from functions and methods in the net package typically implement the net.Error interface, which includes two notable methods:Timeout and Temporary. The Timeout method returns true on Unix based operating systems and Windows if the operating system tells Go that the resource is temporarily unavailable, the call would block, or the connection timed out. We'll touch on time-outs and how you can use them to your advantage a bit later in this chapter. The Temporary method returns true if the error's Timeout function returns true, the function call was interrupted, or there are too many open files on the system, usually because you've exceeded the operating system's resource limit.
```

​	从net包中的函数和方法返回的错误通常实现了`net.Error`接口，其中包括两个重要的方法Timeout和Temporary。Timeout方法在基于Unix的操作系统上返回true，如果操作系统告诉Go资源暂时不可用，调用将阻塞，或连接超时。在本章稍后的部分，我们将讨论超时以及如何利用它们。如果错误的Timeout函数返回true，函数调用被中断，或者系统上有太多打开的文件，通常是因为您超过了操作系统的资源限制，则Temporary方法返回true。

```
Since the functions and methods in the net package return the more general error interface, you'll see the code in this chapter use type assertions to verify you received a net.Error,as in Listing 3-4.
```



​	因为在net包中的函数和方法返回的是更通用的错误接口，所以本章中的代码将使用类型断言来验证接收到的`net.Error`，如清单3-4所示。

```
// err.(net.Error) 断言ok为 true err类型是net.Error。
if nErr, ok := err.(net.Error); ok && !nErr.Temporary() { 
	return err 
}
//清单3-4：断言net.Error以检查错误是否是临时的
```

```
Robust network code won't rely exclusively on the error interface. Rather, it will readily use net.Error's methods, or even dive in further and assert the underlying net.OpError struct, which contains more details about the connection, such as the operation that caused the error, the network type, the source address, and more. I encourage you to read the net.OpError documentation (available at https://golang.org/pkg/net/#OpError/) to learn more about 
specific errors beyond what the net.Error interface provides.
```

​	鲁棒的网络代码不会完全依赖于错误接口。相反，它会随时使用net.Error的方法，甚至深入和断言底层`net.OpError`结构，它包含有关连接的更多细节，例如导致错误的操作、网络类型、源地址等。我鼓励您阅读`net.OpError`文档(可以在 https://golang.org/pkg/net/#OpError/)，了解`net.Error`接口所提供的以外的具体错误的更多信息

### 超时连接`DialTimeout`功能

```
Using the Dial function has one potential problem: you are at the mercy of the operating system to time out each connection attempt. For example, if you use the Dial function in an interactive application and your operating system times out connection attempts after two hours, your application's user may not want to wait that long, much less give your app a five-star rating.
```

使用`Dial`功能有一个潜在的问题:操作系统会让您每次连接尝试都超时。例如，如果您在交互式应用程序中使用`Dial`功能，而您的操作系统在两个小时后断开连接尝试，您的应用程序的用户可能不希望等待那么长时间，更不用说给您的应用程序一个五星评级。

```
To keep your applications predictable and your users happy, it'd be better to control time-outs yourself. For example, you may want to initiate a connection to a low-latency service that responds quickly if it's available. If the service isn't responding, you'll want to time out quickly and move onto the next service.
```

为了保持应用程序的可预测性和用户满意，最好自己控制超时。例如，您可能希望发起一个到低延迟服务的连接，如果服务可用，该服务会快速响应。如果服务没有响应，您将希望快速超时并转移到下一个服务。

```
One solution is to explicitly define a per-connection time-out duration and use the DialTimeout function instead. Listing 3-5 implements this solution
```

一种解决方案是显式定义每个连接超时时间，并使用`DialTimeout`函数代替。清单3-5实现了这个解决方案

```go
// dial_timeout_test.go
// 清单3-5：指定发起TCP连接时的超时时间
package ch03

import (
	"net"
	"syscall"
	"testing"
	"time"
)

// 1
func DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	d := net.Dialer{
		//2
		Control: func(_, addr string, _ syscall.RawConn) error {
			return &net.DNSError{
				Err:         "connection timed out",
				Name:        addr,
				Server:      "127.0.0.1",
				IsTimeout:   true,
				IsTemporary: true,
			}
		},
		Timeout: timeout,
	}
	return d.Dial(network, address)
}
func TestDialTimeout(t *testing.T) {
	// 3
	c, err := DialTimeout("tcp", "10.0.0.1:http", 5*time.Second)
	if err == nil {
		c.Close()
		t.Fatal("connection did not time out")
	}
	//4
	nErr, ok := err.(net.Error)
	if !ok {
		t.Fatal(err)
	}
	//5
	if !nErr.Timeout() {
		t.Fatal("error is not a timeout")
	}
}

```

```
Since the net.DialTimeout function 1 does not give you control of its net.Dialer to mock the dialer's output, you're using our own implementation that matches the signature. Your DialTimeout function overrides the Control function 2 of the net.Dialer to return an error. You're mocking a DNS time-out error.
```

因为`net.DialTimeout`函数1不能让你控制它的网络。要模拟`net.Dialer`的输出，您将使用与签名匹配的我们自己的实现。您的`DialTimeout`功能覆盖网络的控制功能2。`net.Dialer`返回错误。您正在模拟DNS超时错误。

```

Unlike the net.Dial function, the DialTimeout function includes an additional argument, the time-out duration 3. Since the time-out duration is five seconds in this case, the connection attempt will time out if a connection isn't successful within five seconds. In this test, you dial 10.0.0.0, which is a non-routable IP address, meaning your connection attempt assuredly times out. For the test to pass, you need to first use a type assertion to verify you've received a net.Error 4 before you can check its Timeout method 5.If you dial a host that resolves to multiple IP addresses, Go starts a connection race between each IP address, giving the primary IP address a head start. The first connection to succeed persists, and the remaining contenders cancel their connection attempts. If all connections fail or time out, net.DialTimeout returns an error.
```

与`net.Dial`功能不同，`DialTimeout`功能包含一个附加参数，超时时间*3*。在这种情况下，超时时间为5秒，如果连接在5秒内没有成功，则连接尝试将超时。在这个测试中，您(`net.Dial`)拨打10.0.0.0，这是一个不可路由的IP地址，这意味着您的连接尝试肯定会超时。要通过测试，首先需要使用类型断言来验证收到了`net.Error` 4，然后才能检查Timeout方法5。如果您(`net.Dial`)拨打的主机解析到多个IP地址，Go开始每个IP地址之间的连接竞争，让主IP地址领先一步。第一个成功的连接将持续存在，其余的竞争者将取消它们的连接尝试。如果所有连接失败或超时，`net.DialTimeout `返回一个错误。

### 使用带有超时时间的Context来超时连接

```
A more contemporary solution to timing out a connection attempt is to 
use a context from the standard library's context package. A context is an 
object that you can use to send cancellation signals to your asynchronous 
processes. It also allows you to send a cancellation signal after it reaches a 
deadline or after its timer expires.
```



超时连接尝试的一个更现代的解决方案是使用标准库的`context`包中的上下文。context是一个可用于向异步进程发送取消信号的对象。它还允许您在到达截止日期或计时器到期后发送取消信号

```
All cancellable contexts have a corresponding cancel function returned upon instantiation. The cancel function offers increased flexibility since you can optionally cancel the context before the context reaches its deadline. You could also pass along its cancel function to hand off cancellation control to other bits of your code. For example, you could monitor for specific signals from your operating system, such as the one sent to your application when a user presses the ctrl-C key combination, to gracefully abort connection attempts and tear down existing connections before terminating your application.
```

​	所有可取消化上下文在实例化时都返回一个相应的`cancel`函数。`cancel`函数提供了更大的灵活性，因为您可以选择在上下文达到其截止日期前取消上下文。您还可以传递它的取消函数，以将取消控制传递给代码的其他位。例如，您可以监控来自操作系统的特定信号，例如当用户按ctrl-C键组合时发送给应用程序的信号`SIGINT`，以便在终止应用程序之前优雅地终止连接尝试并删除现有连接。

清单3-6说明了一个测试，它使用上下文完成了与对话框超时相同的功能

```go

// 使用具有截止日期的上下文来超时连接尝试
// dial_context_test.go
package ch03

import (
	"context"
	"net"
	"syscall"
	"testing"
	"time"
)

func TestDialContext(t *testing.T) {
	// 1
	dl := time.Now().Add(5 * time.Second)
	// 2
	ctx, cancel := context.WithDeadline(context.Background(), dl)
	// 3
	defer cancel()
	var d net.Dialer // DialContext is a method on a Dialer
	// 4
	d.Control = func(_, _ string, _ syscall.RawConn) error {
		// Sleep long enough to reach the context's deadline.
		time.Sleep(5*time.Second + time.Millisecond)
		return nil
	}
	// 5
	conn, err := d.DialContext(ctx, "tcp", "10.0.0.0:80")

	if err == nil {
		conn.Close()
		t.Fatal("connection did not time out")
	}

	nErr, ok := err.(net.Error)

	if !ok {
		t.Error(err)
	} else {
		if !nErr.Timeout() {
			t.Errorf("error is not a timeout: %v", err)
		}
	}
	// 6
	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("expected deadline exceeded; actual: %v", ctx.Err())
	}
}
```

```
Before you make a connection attempt, you create the context with a deadline of five seconds into the future 1, after which the context will automatically cancel. Next, you create the context and its cancel function by using the context.WithDeadline function 2, setting the deadline in the process. It's good practice to defer the cancel function 3 to make sure the context is garbage collected as soon as possible. Then, you override the dialer's Control function 4 to delay the connection just long enough to make sure you exceed the context's deadline. Finally, you pass in the context as the first argument to the DialContext function 5. The sanity check 6 at the end of the test makes sure that reaching the deadline canceled the context, not an erroneous call to cancel.
```

​	在尝试连接之前，将创建最后期限为5秒的上下文，之后上下文将自动取消。接下来，您可以使用上下文创建上下文及其取消函数。取消截止日期函数2，在过程中设置截止日期。最好推迟取消函数3，以确保上下文尽快被垃圾收集。然后，覆盖拨号器的控制功能4，使连接延迟足够长，以确保超过上下文的截止期限。最后，将上下文作为第一个参数传递给对话框上下文函数5。在测试结束时的理智检查6可以确保在到达截止日期时取消了上下文，而不是一个错误的取消呼叫。

```
As with DialTimeout, if a host resolves to multiple IP addresses, Go starts a connection race between each IP address, giving the primary IP address a head start. The first connection to succeed persists, and the remaining contenders cancel their connection attempts. If all connections fail or the context reaches its deadline, net.Dialer.DialContext returns an error.
```

与`DialTimeout`一样，如果主机解析为多个IP地址，Go开始在每个IP地址之间的连接竞争，使主IP地址头启动。第一个成功的连接持续存在，其余竞争者取消其连接尝试。如果所有连接都失败或上下文达到其截止日期，`net.Dialer.DialContext`将返回一个错误。

### 通过取消该上下文来中止连接

​	**Aborting a Connection by Canceling the Context**

```
Another advantage to using a context is the cancel function itself. You can use it to cancel the connection attempt on demand, without specifying a deadline, as shown in Listing 3-7.
```

使用上下文的另一个优点是` cancel`函数本身。您可以使用它按需取消连接尝试，而不指定截止日期，如清单3-7所示。

```go
// 直接取消上下文以中止连接尝试
// dial_cancel_test.go
package ch03

import (
	"context"
	"net"
	"syscall"
	"testing"
	"time"
)

func TestDialContextCancel(t *testing.T) {
	// 1
	ctx, cancel := context.WithCancel(context.Background())
	sync := make(chan struct{})
	// 2
	go func() {
		defer func() { sync <- struct{}{} }()
		var d net.Dialer
		d.Control = func(_, _ string, _ syscall.RawConn) error {
			time.Sleep(time.Second)
			return nil
		}
		conn, err := d.DialContext(ctx, "tcp", "10.0.0.1:80")
		if err != nil {
			t.Log(err)
			return
		}
		conn.Close()
		t.Error("connection did not time out")
	}()
	// 3
	cancel()
	<-sync
	// 4
	if ctx.Err() != context.Canceled {
		t.Errorf("expected canceled context; actual: %q", ctx.Err())
	}
}
```

```
Instead of creating a context with a deadline and waiting for the deadline to abort the connection attempt, you use context.WithCancel to return a context and a function to cancel the context 1. Since you're manually canceling the context, you create a closure and spin it off in a goroutine to handle the connection attempt 2. Once the dialer is attempting to connect to and handshake with the remote node, you call the cancel function 3to cancel the context. This causes the DialContext method to immediately return with a non-nil error, exiting the goroutine. You can check the context's Err method to make sure the call to cancel was what resulted in the canceled context, as opposed to a deadline in Listing 3-6. In this case, the context's Err method should return a context.Canceled error 4.
```

您可以使用上下文，而不是创建具有截止日期的上下文并等待截止日期终止连接尝试。你使用`context.WithCancel`来返回一个上下文和一个函数来取消该上下文。因为您手动取消了上下文，所以您创建了一个闭包，并在goroutine中分离它来处理连接尝试2。一旦拨号器试图连接到远程节点并与之握手，就调用cancel函数3来取消上下文。这将导致`DialContext`方法立即返回一个非nil错误，退出goroutine。您可以检查上下文的Err方法，以确保对`cancel`的调用是导致被取消上下文的原因，而不是清单3-6中的截止日期。在这种情况下，上下文的Err方法应该返回一个`context.Canceled`错误4。

### 取消多个拨号器

- **Canceling Multiple Dialers** 

```
You can pass the same context to multiple DialContext calls and cancel 
all the calls at the same time by executing the context's cancel function. 
For example, let's assume you need to retrieve a resource via TCP that is 
on several servers. You can asynchronously dial each server, passing each 
dialer the same context. You can then abort the remaining dialers after you 
receive a response from one of the servers
```

您可以将同一上下文传递到多个对话框上下文调用，并通过执行该上下文的`cancel`函数来同时取消所有调用。例如，假设您需要通过多个服务器上的TCP检索资源。您可以异步拨号每个服务器，传递每个拨号器相同的上下文。收到其中一个服务器的响应后，可以中止剩余拨号器

```
In Listing 3-8, you pass the same context to multiple dialers. When you 
receive the first response, you cancel the context and abort the remaining 
dialers
```

在清单3-8中，您可以将相同的上下文传递给多个拨号器。收到第一个响应后，将取消上下文并中止其余的拨号器

```
// 在收到第一个响应后，正在取消所有未完成的拨号器
// dial_fanout_test.go
package ch03

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

func TestDialContextCancelFanOut(t *testing.T) {
	//1
	ctx, cancel := context.WithDeadline(
		context.Background(),
		time.Now().Add(10*time.Second),
	)
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	//2
	go func() {
		// Only accepting a single connection.
		conn, err := listener.Accept()
		if err == nil {
			conn.Close()
		}
	}()
	//3
	dial := func(ctx context.Context, address string, response chan int,
		id int, wg *sync.WaitGroup) {
		defer wg.Done()
		var d net.Dialer
		c, err := d.DialContext(ctx, "tcp", address)
		if err != nil {
			return
		}
		c.Close()
		select {
		case <-ctx.Done():
		case response <- id:
		}
	}
	res := make(chan int)
	var wg sync.WaitGroup
	// 4
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go dial(ctx, listener.Addr().String(), res, i+1, &wg)
	}
	// 5
	response := <-res
	cancel()
	wg.Wait()
	close(res)
	// 6
	if ctx.Err() != context.Canceled {
		t.Errorf("expected canceled context; actual: %s",
			ctx.Err(),
		)
	}
	t.Logf("dialer %d retrieved the resource", response)
}
```

```
You create a context by using context.WithDeadline 1 because you want to have three potential results when checking the context's Err method:context.Canceled, context.DeadlineExceeded, or nil. You expect Err will return the context.Canceled error, since your test aborts the dialers with a call to the cancel function.
```

你通过使用`context.WithDeadline`1创建一个上下文，因为你想在检查上下文的Err方法时有三个潜在的结果:`context.Canceled`， `context.DeadlineExceeded`，或`nil`。您希望Err将返回`context.Canceled`错误，因为您的测试将通过调用cancel函数终止拨号。

```
First, you need a listener. This listener accepts a single connection and closes it after the successful handshake 2. Next, you create your dialers. Since you're spinning up multiple dialers, it makes sense to abstract the dialing code to its own function 3. This anonymous function dials out to the given address by using DialContext. If it succeeds, it sends the dialer's ID across the response channel, provided you haven't yet canceled the context. You spin up multiple dialers by calling dial in separate goroutines using a for loop 4. If dial blocks on the call to DialContext because another dialer won the race, you cancel the context, either by way of the cancel function or the deadline,causing the dial function to exit early. You use a wait group to make sure the test doesn't proceed until all dial goroutines terminate after you cancel the context.
```

首先，你需要一个监听器。这个监听器接受一个连接，并在成功握手2之后关闭它。接下来，创建拨号器。由于您正在轮询多个拨号器，因此将拨号代码抽象为它自己的函数3是有意义的。这个匿名函数使用`DialContext`拨出给定的地址。如果成功，它将通过响应通道发送拨号器的ID，前提是您还没有取消上下文。通过使用for循环4在单独的goroutine中调用dial，可以轮询多个拨号器。如果拨号阻塞对`DialContext`的调用，因为另一个拨号器赢得了比赛，您取消上下文，通过取消功能或截止日期，使拨号功能提前退出。您可以使用一个等待组(WaitGroup)来确保在取消上下文后所有拨号goroutine终止之前，测试不会继续进行。

```
Once the goroutines are running, one will win the race and make a successful connection to the listener. You receive the winning dialer's ID on the res channel 5, then abort the losing dialers by canceling the context. At this point, the call to wg.Wait blocks until the aborted dialer goroutines return. 
Finally, you make sure it was your call to cancel that caused the cancelation of the context 6. Calling cancel does not guarantee that Err will return context.Canceled. The deadline can cancel the context, at which point calls to cancel become a no-op and Err will return context.DeadlineExceeded. In practice, the distinction may not matter to you, but it's there if you need it
```

一旦这些goroutines开始运行，一个goroutines赢得竞争资源，并与`listener`成功建立连接。您在`res channel`通道5上收到获胜拨号器的ID，然后通过取消上下文中止丢失的拨号器。在这一点上，调用`wg.wait` 阻塞，直到终止的拨号器goroutines返回。最后，你要确保是你调用取消导致了上下文的取消。调用cancel并不保证Err将返回`context.Canceled`。截止日期可以取消上下文，此时对cancel的调用变成无操作，Err将返回`context.DeadlineExceeded`。在实践中，这种区别可能对你来说无关紧要，但如果你需要它，它就在那里。

## 实施截止日期

 **Implementing Deadlines**

```
Go's network connection objects allow you to include deadlines for both read and write operations. Deadlines allow you to control how long network connections can remain idle, where no packets traverse the connection. You can control the Read deadline by using the SetReadDeadline method on the connection object, control the Write deadline by using the SetWriteDeadline method, or both by using the SetDeadline method. When a connection reaches its read deadline, all currently blocked and future calls to a network connection's Read method immediately return a time-out error. Likewise, a network connection's Write method returns a time-out error when the connection reaches its write deadline.
```

`Go`的网络连接对象允许你包括读`reader`和写`writer`操作的截止日期。截止时间允许您控制网络连接空闲的时间，即没有数据包通过连接的时间。你可以通过在连接对象上使用`SetReadDeadline`方法来控制读的截止时间，通过`SetWriteDeadline`方法来控制写的截止时间，或者通过`SetDeadline`方法来控制两者。当一个连接达到它的读截止日期时，所有当前阻塞的以及将来对网络连接的`Read`方法的调用都会立即返回超时错误。同样，网络连接的`Write`方法在连接达到写期限时返回超时错误。

```
Go's network connections don't set any deadline for reading and writing operations by default, meaning your network connections may remain idle for a long time. This could prevent you from detecting network failures, like an unplugged cable, in a timely manner, because it's tougher to detect network issues between two nodes when no traffic is in flight.
The server in Listing 3-9 implements a deadline on its connection object
```

默认情况下，Go的网络连接没有设置任何读写操作的截止日期，这意味着你的网络连接可能会在很长一段时间内处于空闲状态。这可以防止您及时检测到网络故障，比如未插的电缆，因为在没有流量的情况下，很难检测到两个节点之间的网络问题。

清单3-9中的服务器为其连接对象实现了一个截止日期

```go
// 由服务器强制执行的最后期限将终止网络连接
// dealine_test.go
package ch03

import (
	"io"
	"net"
	"testing"
	"time"
)

func TestDeadline(t *testing.T) {
	sync := make(chan struct{})
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Log(err)
			return
		}
		defer func() {
			conn.Close()
			close(sync) // read from sync shouldn't block due to early return
		}()
		// 1
		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}
		buf := make([]byte, 1)
		_, err = conn.Read(buf) // blocked until remote node sends data
		nErr, ok := err.(net.Error)
		// 2
		if !ok || !nErr.Timeout() {
			t.Errorf("expected timeout error; actual: %v", err)
		}
		sync <- struct{}{}
		// 3
		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}
		_, err = conn.Read(buf)
		if err != nil {
			t.Error(err)
		}
	}()
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	<-sync
	_, err = conn.Write([]byte("1"))
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 1)
	_, err = conn.Read(buf) // blocked until remote node sends data
	// 4
	if err != io.EOF {
		t.Errorf("expected server termination; actual: %v", err)
	}
}

```

```
Once the server accepts the client's TCP connection, you set the connection's read deadline 1. Since the client won't send data, the call to Read will block until the connection exceeds the read deadline. After five seconds, Read returns an error, which you verify is a time-out 2. Any future reads to the connection object will immediately result in another time-out error. However, you can restore the functionality of the connection object by pushing the deadline forward again 3. After you've done this, a second call to Read succeeds. The server closes its end of the network connection, which initiates the termination process with the client. The client, currently blocked on its Read call, returns io.EOF 4 when the network connection closes.
```

一旦服务器接受了客户端的TCP连接，您就将该连接的读取期限设置为1。因为客户端不会发送数据，所以对Read的调用将被阻塞，直到连接超过读的截止时间。五秒钟后，Read返回一个错误，您验证该错误为超时2。将来对连接对象的任何读取都会立即导致另一个超时错误。但是，您可以通过再次将截止日期往前推来恢复连接对象的功能3。完成这一步后，对Read的第二次调用成功。服务器关闭其网络连接的末端，这将与客户端发起终止进程。当前在其Read调用上被阻塞的客户端。当网络连接关闭时，返回`io.EOF `4。

```
We typically use deadlines to provide a window of time during which the remote node can send data over the network connection. When you read data from the remote node, you push the deadline forward. The remote node sends more data, and you push the deadline forward again, and so on. If you don't hear from the remote node in the allotted time, you can assume that either the remote node is gone and you never received its FIN or that it is idle.
```

​	我们通常使用截止日期来提供一个时间窗口，在此期间远程节点可以通过网络连接发送数据。当从远程节点读取数据时，将把截止日期往前推。远程节点发送更多数据，然后再次将截止日期往前推，以此类推。如果在分配的时间内没有收到远程节点的消息，则可以假设远程节点已经不在并且从未收到其FIN，或者远程节点处于空闲状态。

### 实现一个心跳

​	**Implementing a Heartbeat**

```
For long-running network connections that may experience extended idle periods at the application level, it's wise to implement a heartbeat between nodes to advance the deadline. This allows you to quickly identify network issues and promptly reestablish a connection as opposed to waiting to detect the network error when your application goes to transmit data. In this way, you can help make sure your application always has a good network connection when it needs it.
```

对于长时间运行的网络连接，在应用程序级别可能会经历较长的空闲时间，明智的做法是在节点之间实现心跳，以提前截止时间。这允许您快速识别网络问题并迅速重新建立连接，而不是在应用程序传输数据时等待检测网络错误。通过这种方式，您可以确保应用程序在需要时 始终具有良好的网络连接。

```
For our purposes, a heartbeat is a message sent to the remote side with the intention of eliciting a reply, which we can use to advance the deadline of our network connection. Nodes send these messages at a regular interval, like a heartbeat. Not only is this method portable over various operating systems, but it also makes sure the application using the network connection is responding, since the application implements the heartbeat. Also, this technique tends to play well with firewalls that may block TCP keepalives. We'll discuss keepalives in Chapter 4
```

对于我们的目的，心跳是发送到远程端的消息，目的是引出回复，我们可以使用它来提前网络连接的截止日期。节点会定期发送这些消息，像心跳一样。该方法不仅可以移植到各种操作系统上，而且还可以确保使用网络连接的应用程序正在响应，因为该应用程序实现了心跳。此外，这种技术往往能很好地使用可能阻止TCP保留文件的防火墙。我们将在第4章讨论保活`keepalives`

```
To start, you'll need a bit of code you can run in a goroutine to ping at regular intervals. You don't want to needlessly ping the remote node when you recently received data from it, so you need a way to reset the ping timer. 

Listing 3-10 is a simple implementation from a file named ping.go that meets those requirements
```

首先，您需要一些代码，您可以在一个例行程序中运行，以定期ping。您不希望在最近从远程节点接收数据时不必要地ping该节点，因此需要一种方法来重置ping计时器。

清单3-10是一个名为ping的文件的简单实现。去满足这些要求

```
I use ping and pong messages in my heartbeat examples, where the reception of a ping message—the challenge—tells the receiver it should reply with a pong message—the response. The challenge and response messages are arbitrary. You could use anything you want to here, provided the remote node knows your intention is to elicit its reply.
```

在我的心跳示例中，我使用ping和pong消息，其中ping消息的接收——挑战——告诉接收者它应该用pong消息回应响应。挑战和响应消息是任意的。您可以在这里使用任何您想要的东西，只要远程节点知道您的目的是引出它的应答。

```
// 一种定期对网络连接进行ping检查的功能
// ping.go
1package ch03

import (
	"context"
	"io"
	"time"
)

const defaultPingInterval = 30 * time.Second

func Pinger(ctx context.Context, w io.Writer, reset <-chan time.Duration) {
	var interval time.Duration
	select {
	case <-ctx.Done():
		return
	// 1
	case interval = <-reset: // pulled initial interval off reset channel
	default:
	}
	if interval <= 0 {
		interval = defaultPingInterval
	}
	// 2
	timer := time.NewTimer(interval)
	defer func() {
		if !timer.Stop() {
			<-timer.C
		}
	}()
	for {
		select {
		// 3
		case <-ctx.Done():
			return
		// 4
		case newInterval := <-reset:
			if !timer.Stop() {
				<-timer.C
			}
			if newInterval > 0 {
				interval = newInterval
			}
		// 5
		case <-timer.C:
			if _, err := w.Write([]byte("ping")); err != nil {
				// track and act on consecutive timeouts here
				return
			}
		}
		// 6
		_ = timer.Reset(interval)
	}
}
```

```
The Pinger function writes ping messages to a given writer at regular intervals. Because it's meant to run in a goroutine, Pinger accepts a context as its first argument so you can terminate it and prevent it from leaking. Its remaining arguments include an io.Writer interface and a channel to signal a timer reset. You create a buffered channel and put a duration on it to set the timer's initial interval 1. If the interval isn't greater than zero, you use the default ping interval.
```

Pinger函数定期将ping消息写入给定的写入器(Writer)。因为它是在goroutine中运行的，所以Pinger接受上下文作为它的第一个参数，这样您就可以终止它并防止它泄漏。它的其余参数包括`io.Writer`接口和一个信号定时器重置的通道。您创建一个带缓冲通道，并在其上放置一个持续时间，以设置计时器的初始间隔#1。如果时间间隔不大于零( < 0 )，则使用缺省时间间隔。

```
You initialize the timer to the interval 2 and set up a deferred call to drain the timer's channel to avoid leaking it, if necessary. The endless for loop contains a select statement, where you block until one of three things happens: the context is canceled, a signal to reset the timer is received, or the timer expires. If the context is canceled 3, the function returns, and no further pings will be sent. If the code selects the reset channel 4, you shouldn't send a ping, and the timer resets 6 before iterating on the select statement again.
```

您将计时器初始化为间隔#2，并设置一个延迟调用来耗尽计时器的通道，以避免在必要时泄漏它。无尽的for-loop包含一个select语句，在该语句中，您将阻塞，直到以下三种情况之一发生:取消上下文、接收到重置计时器的信号或计时器过期。如果上下文被取消3，函数将返回，并且不会再发送ping信号。如果代码选择了重置通道4，您不应该发送ping，并且计时器在再次遍历select语句之前重置6。

```
If the timer expires 5, you write a ping message to the writer, and the timer resets before the next iteration. If you wanted, you could use this case to keep track of any consecutive time-outs that occur while writing to the writer. To do this, you could pass in the context's cancel function and call it here if you reach a threshold of consecutive time-outs.
```

如果计时器过期 5，向写入器写入ping消息，计时器在下一次迭代之前重置。如果您需要，可以使用这种情况来跟踪向编写器写入时发生的任何连续超时。为此，可以传入上下文的cancel函数，并在达到连续超时阈值时调用它。

```
Listing 3-11 illustrates how to use the Pinger function introduced in Listing 3-10 by giving it a writer and running it in a goroutine. You can then read pings from the reader at the expected intervals and reset the ping timer with different intervals.
```

清单3-11说明了如何使用清单3-10中介绍的Pinger函数，给它一个写入器并在goroutine中运行。然后，您可以以预期的间隔读取打印，并以不同的间隔重置ping计时器。

```
package ch03

import (
	"context"
	"fmt"
	"io"
	"time"
)

func ExamplePinger() {
	ctx, cancel := context.WithCancel(context.Background())
	r, w := io.Pipe() // in lieu of net.Conn
	done := make(chan struct{})
	// #1 初始化重置器chan 以设置计时器的初始间隔 time.Duration
	resetTimer := make(chan time.Duration, 1)
	resetTimer <- time.Second // initial ping interval
	go func() {
		Pinger(ctx, w, resetTimer)
		close(done)
	}()
	receivePing := func(d time.Duration, r io.Reader) {
		if d >= 0 {
			fmt.Printf("resetting timer (%s)\n", d)
			resetTimer <- d
		}
		now := time.Now()
		buf := make([]byte, 1024)
		n, err := r.Read(buf)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("received %q (%s)\n",
			buf[:n], time.Since(now).Round(100*time.Millisecond))
	}
	// #2 您将计时器初始化为间隔
	for i, v := range []int64{0, 200, 300, 0, -1, -1, -1} {
		fmt.Printf("Run %d:\n", i+1)
		receivePing(time.Duration(v)*time.Millisecond, r)
	}
	cancel()
	<-done // ensures the pinger exits after canceling the context
	// Output:
	// #3
	// Run 1:
	// resetting timer (0s)
	// received "ping" (1s)
	// #4
	// Run 2:
	// resetting timer (200ms)
	// received "ping" (200ms)
	// #5
	// Run 3:
	// resetting timer (300ms)
	// received "ping" (300ms)
	// #6
	// Run 4:
	// resetting timer (0s)
	// received "ping" (300ms)
	// #7
	// Run 5:
	// received "ping" (300ms)
	// Run 6:
	// received "ping" (300ms)
	// Run 7:
	// received "ping" (300ms)
}
```

Listing3-11 注解：

```
In this example, you create a buffered channel 1 that you'll use to signal a reset of the Pinger's timer. You put an initial ping interval of one second on the resetTimer channel before passing the channel to the Pinger function. You'll use this duration to initialize the Pinger's timer and dictate when to write the ping message to the writer.
```


在本例中，您创建了一个缓冲通道1，您将使用该通道发出Pinger计时器重置的信号。在将该通道传递给Pinger函数之前，在`resetTimer`通道上设置一个1秒的初始ping间隔。您将使用此持续时间初始化Pinger的计时器，并指定何时将ping消息写入写入器

```
You run through a series of millisecond durations in a loop 2, passing each to the receivePing function. This function resets the ping timer to the given duration and then waits to receive the ping message on the given reader. Finally, it prints to stdout the time it takes to receive the ping message. Go checks stdout against the expected output in the example.
```

在循环2中运行一系列毫秒的持续时间，并将每个毫秒传递给接收函数。这个函数将`ping`定时器重置为给定的持续时间，然后在给定的读取器上等待接收`ping`消息。最后，它将接收`ping`消息所需的时间打印到控制台`stdout`。根据示例中的预期输出检查控制台的输出结果`stdout`。

```
During the first iteration 3, you pass in a duration of zero, which tells the Pinger to reset its timer by using the previous duration—one second in this example. As expected, the reader receives the ping message after one second. The second iteration 4 resets the ping timer to 200 ms. Once this expires, the reader receives the ping message. The third run resets the ping timer to 300 ms 5, and the ping arrives at the 300 ms mark.
```

在第一次迭代中 3，传入一个持续时间为0的值，它告诉`Pinger`使用之前的持续时间(在本例中为1秒)重置其计时器。正如所料，读入器在一秒后收到`ping`消息。第二次迭代4将ping定时器重置为200毫秒。一旦过期，读入器就会收到`ping`消息。第三次运行将ping定时器重置为`300 ms` 5，并且`ping`到达`300 ms`标记。

```
You pass in a zero duration for run 4 #6, preserving the 300 ms ping timer from the previous run. I find the technique of using zero durations to mean “use the previous timer duration” useful because I do not need to keep track of the initial ping timer duration. I can simply initialize the timer with the duration I want to use for the remainder of the TCP session and reset the timer by passing in a zero duration every time I need to preempt the transmission of the next ping message. Changing the ping timer duration in the future involves the modification of a single line as opposed to every place I send on the resetTimer channel.
```

您为运行4 #6传递了一个0持续时间，保留了上一次运行的`300`毫秒的`ping`计时器。我发现使用零持续时间表示“使用前一个计时器持续时间”的技术很有用，因为我不需要跟踪初始`ping`计时器持续时间。我可以简单地用我希望用于其余TCP会话的持续时间初始化定时器，并通过在每次需要抢占下一个ping消息的传输时传入一个零持续时间来重置定时器。改变未来的`ping`定时器持续时间涉及修改单行，而不是我在`resetTimer`通道上发送的每个位置。

```
Runs 5 to 7 #7 simply listen for incoming pings without resetting the ping timer. As expected, the reader receives a ping at 300 ms intervals for the last three runs.
```

运行5到7只是监听传入的ping，而不重置ping计时器。正如预期的那样，在最后三次运行中，读取器以`300ms`的间隔收到一个ping。

```
With Listing 3-10 saved to a file named ping.go and Listing 3-11 saved to 
a file named ping_example_test.go, you can run the example by executing the 
following
```

将清单3-10保存到名为ping.go的文件，将清单3-11保存为名为ping_example_test.go的文件，可以通过执行以下操作来运行该示例

```shell
 go test ping.go ping_example_test.go
```

### 通过使用心跳来推进截止日期

**Advancing the Deadline by Using the Heartbeat**

```
Each side of a network connection could use a Pinger to advance its deadline if the other side becomes idle, whereas the previous examples showed only a single side using a Pinger. When either node receives data on the network connection, its ping timer should reset to stop the delivery of an unnecessary ping.
Listing 3-12 is a new file named ping_test.go that shows how you can use incoming messages to advance the deadline
```

如果另一侧空闲，网络连接的每一边都可以使用Pinger提前其最后期限，而前面的例子只显示了使用Pinger的单一面。当任何一个节点接收到网络连接上的数据时，其ping计时器应重置以停止传递不必要的ping。清单3-12是一个名为ping_test.go的新文件，它显示了如何使用传入的消息提前发送截止日期

如果另一侧空闲，网络连接的每一边都可以使用Pinger提前其最后期限，而前面的例子只显示了使用Pinger的单一面。当任何一个节点接收到网络连接上的数据时，其ping计时器应重置以停止传递不必要的ping。清单3-12是一个名为ping_test.go的新文件，它显示了如何使用传入的消息提前发送截止日期

```go
// 清单3-12：接收数据提前了截止日期
// ping_test.go
package ch03

import (
	"context"
	"io"
	"net"
	"testing"
	"time"
)

func TestPingerAdvanceDeadline(t *testing.T) {
	done := make(chan struct{})
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	begin := time.Now()
	go func() {
		defer func() { close(done) }()
		conn, err := listener.Accept()
		if err != nil {
			t.Log(err)
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer func() {
			cancel()
			conn.Close()
		}()
		resetTimer := make(chan time.Duration, 1)
		resetTimer <- time.Second
		go Pinger(ctx, conn, resetTimer)
		// 1
		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			t.Logf("[%s] %s",
				time.Since(begin).Truncate(time.Second), buf[:n])
			// 2
			resetTimer <- 0
			// 3
			err = conn.SetDeadline(time.Now().Add(5 * time.Second))
			if err != nil {
				t.Error(err)
				return
			}
		}
	}()
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	buf := make([]byte, 1024)
	// 4
	for i := 0; i < 4; i++ { // read up to four pings
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])
	}
	// 5
	_, err = conn.Write([]byte("PONG!!!")) // should reset the ping timer
	if err != nil {
		t.Fatal(err)
	}
	// 6
	for i := 0; i < 4; i++ { // read up to four more pings
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				t.Fatal(err)
			}
			break
		}
		t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])
	}
	<-done
	end := time.Since(begin).Truncate(time.Second)
	t.Logf("[%s] done", end)
	if end != 79*time.Second {
		t.Fatalf("expected EOF at 9 seconds; actual %s", end)
	}
}
```

```
You start a listener that accepts a connection, spins off a Pinger set to ping every second, and sets the initial deadline to five seconds 1. From a client's perspective, it receives four pings followed by an io.EOF when the server reaches its deadline and terminates its side of the connection. However, a client can advance the server's deadline by sending the server data 5 before the server reaches its deadline.
```



您启动一个接受连接的监听器，将Pinger设置为每秒ping一次，并将初始截止日期设置为5秒 1。从客户端的角度来看，当服务器到达其截止日期并终止其连接端的时候，它将接收4个ping，然后是一个`io.EOF`。然而，客户端可以通过在服务器到达其截止日期之前发送服务器数据 5 来提前服务器的截止日期。

```
If the server reads data from its connection, it can be confident the network connection is still good. Therefore, it can inform the Pinger to reset 2 its timer and push the connection's deadline forward 3. To preempt the termination of the socket, the client listens for four ping messages 4 from the server before sending an emphatic pong message 5. This should buy the client five more seconds until the server reaches its deadline. The client reads four more pings 6 and then waits for the inevitable. You check that a total of nine seconds 7 has elapsed by the time the server terminates the connection, indicating the client's pong successfully triggered the reset of the ping timer
```

如果服务器从它的连接中读取数据，它可以确信网络连接仍然是好的。因此，它可以通知Pinger重置它的定时器，并将连接的截止时间往前推。为了抢占套接字的终止，客户端在发送一个强调的pong消息5之前，监听来自服务器的4条ping消息4。这应该能为客户端多争取5秒，直到服务器到达其截止日期。客户端读取另外四个ping  6，然后等待不可避免的结果。您检查服务器终止连接的时间总共已经过去了9秒 7，这表明客户端pong成功触发了ping计时器的重置

```
In practice, this method of advancing the ping timer cuts down on the 
consumption of bandwidth by unnecessary pings. There is rarely a need to 
challenge the remote side of a network connection if you just received data 
on the connection.
```

实际上，这种推进ping计时器的方法减少了不必要的脉冲的带宽消耗。如果您刚刚收到该连接上的数据，则很少需要挑战(challenge)网络连接的远程端。

```
The strings "ping" and "pong" are arbitrary. You could use smaller payloads,such as a single byte, for the same purpose, provided both sides of the network connection agree upon what values constitute a ping and a pong.
```

字符串“ping”和“pong”是任意的。为了达到相同的目的，您可以使用更小的有效负载，例如单个字节，只要网络连接的双方对构成ping和pong的值达成一致。

### 您的经验

- 总结

**What You've Learned**

```
We covered a lot of ground in this chapter. We started with a dive into TCP's handshake, sequences, and acknowledgments, the sliding window, and connection terminations. Then, we covered the process of establishing TCP connections using Go's standard library. We talked about temporary errors, time-outs, listening for incoming connections, and dialing remote services. Finally, we covered techniques to help you detect and timely correct network integrity issues.
```

我们在这章中涵盖了很多内容。我们首先深入了解TCP的握手、序列和确认、滑动窗口和连接终端。然后，我们介绍了使用Go的标准库建立TCP连接的过程。我们讨论了临时错误、超时、监听传入的连接和(client dial)拨号远程服务。最后，我们介绍了一些可以帮助您检测和及时纠正网络完整性问题的技术。

```
I strongly recommend picking up Practical Packet Analysis by Chris Sanders (No Starch Press, 2017) and installing Wireshark. Manipulating your network code and seeing how it affects TCP traffic in Wireshark is a fantastic way to gain a deeper understanding of both TCP and Go's networking packages. The next chapter covers sending and receiving data over TCP connections. Wireshark will help you gain a deeper understanding of data you send, including each payload's effects on the sliding window. Familiarizing yourself with it now will pay dividends
```

我强烈推荐Chris Sanders的practice Packet Analysis (No Starch Press, 2017)并安装Wireshark。在Wireshark中操作你的网络代码，看看它是如何影响TCP流量的，这是一个很好的方法，可以更深入地理解TCP和Go的网络包。下一章将介绍通过TCP连接发送和接收数据。Wireshark将帮助你更深入地理解你发送的数据，包括每个有效载荷对滑动窗口的影响。现在熟悉它会有好处的

## 第4章

**SENDING TCP DATA**  发送`TCP`数据

```
Now that you know how to properly establish and gracefully terminate TCP connections in Go, it's time to put that knowledge to use by transmitting data. This chapter covers various techniques for sending and receiving data over a network using TCP
```



现在您已经知道了如何在Go中正确地建立和优雅地终止TCP连接，现在可以通过传输数据来使用这些知识。本章介绍了使用TCP在网络上发送和接收数据的各种技术

```
We'll talk about the most common methods of reading data from network connections. You'll create a simple messaging protocol that allows you to transmit dynamically sized payloads between nodes. You'll then explore the networking possibilities afforded by the net.Conn interface. The chapter concludes with a deeper dive into the TCPConn object and insidious TCP networking problems that Go developers may experience.
```

我们将讨论从网络连接中读取数据的最常见方法。您将创建一个简单的消息传递协议，允许您在节点之间传输动态大小的有效负载。然后，您将探索`network`提供的`net.Conn`接口。本章以更深入地了解`TCPConn`对象和Go开发人员可能遇到的潜在的`TCP`网络问题结束。



### 使用`net.Conn`接口

**Using the net.Conn Interface**

```
Most of the network code in this book uses Go's net.Conn interface whenever possible, because it provides the functionality we need for most cases. You can write powerful network code using the net.Conn interface without having to assert its underlying type, ensuring your code is compatible across operating systems and allowing you to write more robust tests. (You will learn how to access net.Conn's underlying type to use its more advanced methods later in this chapter.) The methods available on net.Conn cover most use cases.
```



本书中的大部分网络代码都使用了Go的`net.Conn`接口，因为它提供了大多数情况下所需要的功能。您可以使用`net.Conn`接口编写功能强大的网络代码，而不必断言其底层类型, 从而确保您的代码跨操作系统兼容，并允许您编写鲁棒的测试。(在本章的后面，您将学习如何访问`net.Conn`的底层类型来使用它的更高级的方法)`net.Conn`上可用的方法涵盖了大多数用例。

```
The two most useful net.Conn methods are Read and Write. These methods implement the io.Reader and io.Writer interfaces, respectively, which are ubiquitous in the Go standard library and ecosystem. As a result, you can leverage the vast amounts of code written for those interfaces to create incredibly powerful network applications
```

两个最有用的`net.Conn`方法是`Read`和`Write`。这些方法实现`io.Reader`和`io.Writer`接口，它们在Go标准库和生态系统中无处不在。因此，您可以利用为这些接口编写的大量代码来创建功能极其强大的网络应用程序

```
You use net.Conn's Close method to close the network connection. This method will return nil if the connection successfully closed or an error otherwise. The SetReadDeadline and SetWriteDeadline methods, which accept a time.Time object, set the absolute time after which reads and writes on the network connection will return an error. The SetDeadline method sets both the read and write deadlines at the same time. As discussed in “Implementing Deadlines” on page 62, deadlines allow you to control how long a network connection may remain idle and allow for timely detection of network connectivity problems.
```

使用`net.Conn`的`Close`方法关闭网络的连接，此方法将返回nil，否则将返回错误。`SetReadDeadline`和`SetWriteDeadline`方法,它接受一个`time.Time`对象，设置网络连接读写的绝对时间将返回错误。`SetDeadline`方法可同时设置读取和写取截止日期。正如第62页“实现截止日期”中所讨论的，截止日期允许您控制网络连接的空闲时间，并允许及时检测到网络连接问题。

### 发送和接收数据

**Sending and Receiving Data**

```
Reading data from a network connection and writing data to it is no different from reading and writing to a file object, since net.Conn implements the io.ReadWriteCloser interface used to read and write to files. In this section, you'll first learn how to read data into a fixed-size buffer. Next, you'll learn how to use bufio.Scanner to read data from a network connection until it encounters a specific delimiter. You'll then explore TLV, an encoding method that enables you to define a basic protocol to dynamically allocate buffers for varying payload sizes. Finally, you'll see how to handle errors when reading from and writing to network connections
```



从网络连接读取数据并向其写入数据与读取和写入文件对象没有什么不同，因为`net.Conn`实现了`io.ReadWriteCloser`接口，用于读取和写入文件。在本节中，您将首先学习如何将数据读入固定大小的缓冲区。接下来，您将学习如何使用`bufio.Scanner`从网络连接读取数据，直到遇到特定的分隔符为止。然后将研究**TLV**，这是一种编码方法，使您能够定义基本协议来为不同的有效负载大小动态分配缓冲区。最后，您将看到在读取和写入网络连接时如何处理错误

```
TLV  type length value  thrift 协议就是这样的
gRpc 给每一个字段编号再使用的bit为来编码字段，压缩很高
```

#### **Reading Data into a Fixed Buffer**

```
TCP connections in Go implement the io.Reader interface, which allows you to read data from the network connection. To read data from a network connection, you need to provide a buffer for the network connection's Read method to fill.
```

Go中的TCP连接实现了`io.Reader`接口，它允许你从网络连接中读取数据。要从网络连接读取数据，需要为网络连接的`Read`方法提供一个缓冲区来填充数据。

```
The Read method will populate the buffer to its capacity if there is enough data in the connection's receive buffer. If there are fewer bytes in the receive buffer than the capacity of the buffer you provide, Read will populate the given buffer with the data and return instead of waiting for more data to arrive. In other words, Read is not guaranteed to fill your buffer to capacity before it returns. Listing 4-1 demonstrates the process of reading data from a network connection into a byte slice.
```

​	如果连接的接收缓冲区中有足够的数据，则`Read`方法将填充缓冲区到其容量。如果接收缓冲区中的字节少于您提供的缓冲区容量，`Read`将使用数据填充给定缓冲区并返回，而不是等待更多数据到达。换句话说，`Read`不能保证在返回之前将缓冲区填满。

清单4-1演示了将数据从网络连接读入字节片的过程。

```
// 通过网络连接接收数据
package ch04

import (
	"crypto/rand"
	"io"
	"net"
	"testing"
)

func TestReadIntoBuffer(t *testing.T) {
	t.Log("TestRead")
	//1  定义一个16 MB的切片payload
	payload := make([]byte, 1<<24) // 16 MB
	_, err := rand.Read(payload)   // generate a random payload
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Log(err)
			return
		}
		defer conn.Close()
		// 2 将随机读取的内容写入到conn中
		_, err = conn.Write(payload)
		if err != nil {
			t.Error(err)
		}
	}()
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	// 3 创建一个 512 KB的切片 buf
	buf := make([]byte, 1<<19) // 512 KB
	for {
		//4 从网络中读取数据到切片buf中
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				t.Error(err)
			}
			break
		}
		// 打印切片中的字节个数
		t.Logf("buf:%s\n", string(buf[:n]))
		t.Logf("read %d bytes", n) // buf[:n] is the data read from conn
	}
	conn.Close()
}

// go test -v ./read_test.go
```

```
You need something for the client to read, so you create a 16MB payload of random data 1—more data than the client can read in its chosen buffer size of 512KB 3 so that it will make at least a few iterations around its for loop. It's perfectly acceptable to use a larger buffer or a smaller payload and read the entirety of the payload in a single call to Read. Go correctly processes the data regardless of the payload and receive buffer sizes. 

You then spin up the listener and create a goroutine to listen for incoming connections. Onceaccepted, the server writes the entire payload to the network connection 2. The client then reads up to the first 512KB from the connection 4 before continuing around the loop. The client continues to read up to 512KB at a time until either an error occurs or the client reads the entire 16MB payload.
```

你需要一些东西让客户端去读，所以你创建了一个16MB的随机数据的payload(有效负载) 1 比客户端在512KB的缓冲区中所能读到的更多的数据，这样它就会围绕它的for循环进行至少几次迭代。使用较大的缓冲区或较小的payload并在对read的单个调用中读取整个payload是完全可以接受的。不管payload和接收缓冲区大小如何，Go都正确处理数据。

然后启动监听器并创建一个goroutine来监听传入的连接。`Onceaccepted`，服务器将整个有效负载写入网络连接 2。然后客户端从连接 4  中读取第一个512KB，然后继续循环。客户端每次继续读取最多512KB，直到出现错误或客户端读取整个16MB负载。

### 使用扫描仪分隔读取

**Delimited Reading by Using a Scanner**

```
Reading data from a network connection by using the method I just showed means your code needs to make sense of the data it receives. Since TCP is a stream oriented protocol, a client can receive a stream of bytes across many packets. Unlike sentences, binary data doesn't include inherent punctuation that tells you where one message starts and stops.
```

​	使用我刚才展示的方法从网络连接读取数据意味着您的代码需要理解它接收到的数据。由于TCP是一个面向流的协议，客户端可以接收跨多个包的字节流。与句子不同，二进制数据不包括固有的标点符号来告诉您消息的开始和结束位置。

```
If, for example, your code is reading a series of email messages from a server, your code will have to inspect each byte for delimiters indicating the boundaries of each message in the stream of bytes. Alternatively, your client may have an established protocol with the server whereby the server sends a fixed number of bytes to indicate the payload size the server will send next. Your code can then use this size to create an appropriate buffer for the payload. You'll see an example of this technique a little later in this chapter.
```

例如，如果您的代码正在从服务器读取一系列电子邮件消息，那么您的代码将必须检查每个字节中指示字节流中每个消息边界的分隔符。或者，您的客户端可能有一个与服务器建立的协议，服务器通过该协议发送固定数量的字节来指示服务器下一步将发送的有效载荷大小。然后，您的代码可以使用这个大小为有效负载创建适当的缓冲区。在本章稍后的部分，您将看到这种技术的一个示例。

```
However, if you choose to use a delimiter to indicate the end of one message and the beginning of another, writing code to handle edge cases isn't so simple. For example, you may read 1KB of data from a single Read on the network connection and find that it contains two delimiters. This indicates that you have two complete messages, but you don't have enough information about the chunk of data following the second delimiter to know whether it is also a complete message. If you read another 1KB of data and find no delimiters, you can conclude that this entire block of data is a continuation of the last message in the previous 1KB you read. But what if you read 1KB of nothing but delimiters?
```

但是，如果选择使用分隔符来表示一条消息的结束和另一条消息的开始，那么编写处理边界情况的代码就不是那么简单了。例如，您可以从网络连接上的单个read读取1KB的数据，并发现它包含两个分隔符。这表明您有两个完整的消息，但是您没有关于第二个分隔符后面的数据块的足够信息来知道它是否也是一个完整的消息。如果您读取另一个1KB的数据并没有发现分隔符，那么您可以得出结论，这整个数据块是您读取的前一个1KB中的最后一条消息的延续。但是如果只读取1KB的分隔符怎么办?

```
If this is starting to sound a bit complex, it's because you must account for data across multiple Read calls and handle any errors along the way. Anytime you're tempted to roll your own solution to such a problem, check the standard library to see if a tried-and-true implementation already exists. In this case, bufio.Scanner does what you need.The bufio.Scanner is a convenient bit of code in Go's standard library that allows you to read delimited data. The Scanner accepts an io.Reader as its input. Since net.Conn has a Read method that implements the io.Reader interface, you can use the Scanner to easily read delimited data from a network connection. Listing 4-2 sets up a listener to serve up delimited data for later parsing by bufio.Scanner.
```

如果这听起来有点复杂，那是因为您必须考虑跨多个Read调用的数据，并在此过程中处理任何错误。当您试图用自己的解决方案来解决此类问题时，请检查标准库，看看是否已经存在经过验证的实现。在这种情况下，`bufio.Scanner`你需要的。`bufio.Scanner`是Go标准库中的一个方便的代码，它允许你读取带分隔符的数据。扫描仪接受`io.Reader`作为其输入值。因为 `net.Conn`有一个读取的方法来实现`io.Reader`接口，您可以使用`Scanner`轻松地从网络连接读取分隔的数据。清单4-2设置了一个监听器(listener)来提供带分隔符的数据，以便稍后由`bufio.Scanner`进行解析。

```
// 使用bufio.Scanner从网络中读取以空格分隔的文本
// 创建一个用来提供恒定有效载荷的测试(scanner_test.go)
package ch04

import (
	"bufio"
	"net"
	"reflect"
	"testing"
)

// payload 它的作用就是提供有效载荷
const payload = "The bigger the interface, the weaker the abstraction."

func TestScanner(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		_, err = conn.Write([]byte(payload))
		if err != nil {
			t.Error(err)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	// 创建一个bufio.Scanner,从网络连接读取数据
	scanner := bufio.NewScanner(conn)
	scanner.Split(bufio.ScanWords)
	var words []string
	//2 sacnner从连接读取数据
	for scanner.Scan() {
		//3 将数据块作为字符串、单个单词和相邻的标点符号返回
		words = append(words, scanner.Text())
	}
	err = scanner.Err()
	if err != nil {
		t.Error(err)
	}
	expected := []string{"The", "bigger", "the", "interface,", "the",
		"weaker", "the", "abstraction."}
	if !reflect.DeepEqual(words, expected) {
		t.Fatal("inaccurate scanned word list")
	}
	//4 打印扫描到的单词
	t.Logf("Scanned words: %#v", words)
}

```



```
This listener should look familiar by now. All it's meant to do is serve up 
the payload. 
Listing 4-3 uses bufio.Scanner to read a string from the network, splitting each chunk by whitespace
```

这个`listener`现在看起来应该很熟悉了。它的作用就是提供有效载荷。

清单4-3 使用`bufio.Scanner`从网络中读取以空格分隔的文本

```
Since you know you're reading a string from the server, you start by creating a bufio.Scanner that reads from the network connection. By default, the scanner will split data read from the network connection when it encounters a newline character (\n) in the stream of data. Instead, you elect to have the scanner delimit the input at the end of each word by using bufio.ScanWords, which will split the data when it encounters a word border, such as whitespace or sentence-terminating punctuation.
```

由于您知道您正在从服务器读取字符串，首先创建一个`bufio.Scanner`从网络连接读取数据。默认情况下，当扫描器在数据流中遇到换行符(\n)时，它将拆分从网络连接中读取的数据。相反，您选择让扫描器使用`bufio.ScanWords`的末尾分隔输入。扫描字，它将在遇到单词边框时分隔数据，如空白格或结束句子的标点符号。

```
You keep reading data from the scanner as long as it tells you it's read data from the connection 2. Every call to Scan can result in multiple calls to the network connection's Read method until the scanner finds its delimiter or reads an error from the connection. It hides the complexity of searching for a delimiter across one or more reads from the network connection and returning the resulting messages.
```

只要扫描器告诉你它从连接读取数据，你就一直从扫描器读取数据。对 `Scan`的每次调用都可能导致对网络连接的`Read`方法的多次调用，直到扫描程序找到其分隔符或从连接中读取一个错误。它隐藏了跨网络连接的一次或多次读取搜索分隔符并返回结果消息的复杂性。

```
The call to the scanner's Text method returns the chunk of data as a string a single word and adjacent punctuation, in this case that it just read from the network connection 3. The code continues to iterate around the for loop until the scanner receives an io.EOF or other error from the network connection. If it's the latter, the scanner's Err method will return a non-nil error. You can view the scanned words 4 by adding the -v flag to the go test command.
```

对扫描器的`scanner.Text()`方法的调用将数据块作为字符串、单个单词和相邻的标点符号返回，在本例中，它只是从网络连接3中读取。代码继续遍历for循环，直到扫描器从网络连接接收到`io.EOF`或其他错误。如果是后者，扫描器的`scanner.Err()`方法将返回一个非nil错误。你可以通过在go test命令中添加-v标志来查看扫描到的单词4。

### 动态分配缓冲区的大小

**Dynamically Allocating the Buffer Size**

```
You can read data of variable length from a network connection, provided that both the sender and receiver have agreed on a protocol for doing so. The type-length-value (TLV) encoding scheme is a good option. TLV encoding uses a fixed number of bytes to represent the type of data, a fixed number of bytes to represent the value size, and a variable number of bytes to represent the value itself. Our implementation uses a 5-byte header: 1 byte for the type and 4 bytes for the length. The TLV encoding scheme allows you to send a type as a series of bytes to a remote node and constitute the same type on the remote node from the series of bytes
Listing 4-4 defines the types that our TLV encoding protocol will accept
```

您可以从网络连接读取可变长度的数据，前提是发送方和接收方都使用相同的通讯协议。类型长度值 即type-length-value(TLV)编码方案是一个不错的选择。TLV编码使用固定的字节数来表示数据的类型，使用固定的字节数来表示值的大小，以及使用可变的字节数来表示值本身。我们的实现使用一个5字节的头：类型为1字节，长度为4字节。TLV编码方案允许您将类型作为一系列字节形式发送到远程节点，并从字节系列在远程节点上构成相同的类型

清单4-4定义了TLV编码协议将接受的类型。

```
package ch04

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

/*
	The message struct implements a simple protocol (types.go).
	这个消息结构实现了一个简单的协议
*/
// 创建二进制文件类型
const (
	// 消息类型
	BinaryType uint8 = iota + 1
	// uint8 = iota + 1 字符串类型
	StringType
	//最大的有效载荷 即值的最大长度
	MaxPayloadSize uint32 = 10 << 20 // 10 MB
)

var ErrMaxPayloadSize = errors.New("maximum payload size exceeded")

// 接口定义
type Payload interface {
	fmt.Stringer
	io.ReaderFrom
	io.WriterTo
	Bytes() []byte
}

// Binary 为一个字节切片
type Binary []byte

// 实现Bytes接口,将自己返回
func (m Binary) Bytes() []byte { return m }

// 实现String接口，将自己转为string返回
func (m Binary) String() string { return string(m) }

// 实现WriteTo接口
func (m Binary) WriteTo(w io.Writer) (int64, error) {
	// 调用 Write方法写入 1 byte
	err := binary.Write(w, binary.BigEndian, BinaryType) // 1-byte type
	if err != nil {
		return 0, err
	}
	var n int64 = 1
	// 它将二进制文件的4字节长度写入
	// BigEndian是ByteOrder的大端实现
	err = binary.Write(w, binary.BigEndian, uint32(len(m))) // 4-byte size
	if err != nil {
		return n, err
	}
	n += 4
	// 写入二进制值本身
	o, err := w.Write(m) // payload
	return n + int64(o), err
}

```



```
You start by creating constants to represent each type you will define. In this example, you will create a BinaryType 1 and a StringType 2. After digesting the implementation details of each type, you should be able to create types that fit your needs. For security purposes that we'll discuss in just a moment, you must define a maximum payload size 3.
```

首先可以创建常量来表示您将定义的每个类型。在本示例中，您将创建一个二进制类型即`BinaryType`和一个字符串类型即`StringType`。在消化了每种类型的实现详细信息之后，您应该能够创建适合您需要的类型。出于安全考虑，我们将稍后讨论，您必须定义最大有效载荷大小。

PS: 应该加一个checkSum校验值

```
You also define an interface named Payload 4 that describes the methods each type must implement. Each type must have the following methods: Bytes, String, ReadFrom, and WriteTo. The io.ReaderFrom and io.WriterTo interfaces allow your types to read from readers and write to writers, respectively. You have some flexibility in this regard. You could just as easily make the Payload implement the encoding.BinaryMarshaler interface to marshal itself to a byte slice and the encoding.BinaryUnmarshaler interface to unmarshal itself from a byte slice. But the byte slice is one level removed from the network connection, so you'll keep the Payload interface as is. Besides, you'll use the binary encoding interfaces in the next chapter
You now have the foundation built to create TLV-based types. Listing 4-5 
details the first type, Binary
```

您还定义了一个名为Payload的接口，它描述了每种类型必须实现的方法。每种类型必须有以下方法:`Bytes`、`String`、`ReadFrom`和`WriteTo`。`io.ReaderFrom`和`io.WriterTo`接口允许类型分别从读取器中读取数据和写入器中写入数据。在这方面你可以有一些灵活性。您也可以轻松地让Payload实现`encoding.BinaryMarshaler`接口将PayLoad结构体序列化为字节切片和`encoding.BinaryUnmarshaler`接口反序列化字节切片为PayLoad结构体。但是字节片从网络连接中删除了一个级别，所以您将保持Payload接口不变。此外，您将在下一章中使用二进制编码接口

现在可以创建基于TLV的类型。清单4-5详细说明了第一种类型，二进制文件





```
The Binary type is a byte slice; therefore, its Bytes method simply returns itself. Its String method casts itself as a string before returning. The WriteTo method accepts an io.Writer and returns the number of bytes written to the writer and an error interface 4. The WriteTo method first writes the 1-byte type to the writer. It then writes the 4-byte length of the Binary to the writer 6. Finally, it writes the Binary value itself 7.
```

`Binary type` 是一个字节切片,因此，它的`Bytes`方法只是返回自己。它的`String`方法在返回将字符切片强制转换为一个字符串，`WriteTo`方法接受一个`io.Writer`并返回写入写入器(writer)的字节数和错误接口。`WriteTo`方法首先将1字节类型写入写入器(writer)，然后，它将二进制文件的4字节长度写入写入器(writer)，最后，它写入二进制值本身。

清单4-6概括了二进制类型及其ReadFrom方法。

```go
func (m *Binary) ReadFrom(r io.Reader) (int64, error) {
	var typ uint8
	// eadFrom方法将1个字节从读取器读入类型变量 typ
	err := binary.Read(r, binary.BigEndian, &typ) // 1-byte type
	if err != nil {
	return 0, err
	}
	var n int64 = 1
	// 验证该类型是否为 BinaryType类型
	if typ != BinaryType {
		return n, errors.New("invalid Binary")
 }
 var size uint32
 // 将接下来的4个字节读入`size`变量
 err = binary.Read(r, binary.BigEndian, &size) // 4-byte size
 if err != nil {
 return n, err
 }
 n += 4
 // 防止DOS攻击耗尽主机内存
 if size > MaxPayloadSize {
 	return n, ErrMaxPayloadSize
 }
 // 创建一个新的切片
 *m = make([]byte, size)
 // 填充二进制字节切片
 o, err := r.Read(*m) // payload
 return n + int64(o), err
}
```



```
The ReadFrom method reads 1 byte from the reader into the type variable. It next verifies that the type is BinaryType before proceeding. Then it reads  the next 4 bytes into the size variable, which sizes the new Binary byte slice . Finally, it populates the Binary byte slice .
Notice that you enforce a maximum payload size 4. This is because the 
4-byte integer you use to designate the payload size has a maximum value of 
4,294,967,295, indicating a payload of over 4GB. With such a large payload 
size, it would be easy for a malicious actor to perform a denial-of-service attack 
that exhausts all the available random access memory (RAM) on your computer. 
Keeping the maximum payload size reasonable makes memory exhaustion 
attacks harder to execute.
Listing 4-7 introduces the String type, which, like Binary, implements 
the Payload interface.
```

`ReadFrom`方法从读取器读取1个字节到类型变量中。接下来验证该类型是否为`BinaryType`，然后它将接下来的4个字节读入`size`变量，该变量调整新二进制字节切片的大小，最后，它填充二进制字节切片

```
Notice that you enforce a maximum payload size 4. This is because the 4-byte integer you use to designate the payload size has a maximum value of 4,294,967,295, indicating a payload of over 4GB. With such a large payload size, it would be easy for a malicious actor to perform a denial-of-service attack that exhausts all the available random access memory (RAM) on your computer. Keeping the maximum payload size reasonable makes memory exhaustion attacks harder to execute.
Listing 4-7 introduces the String type, which, like Binary, implements 
the Payload interface
```

请注意，强制执行最大有效负载大小。这是因为您用来指定有效负载大小的4字节整数的最大值为4,294,967,295，表示有效负载超过4GB。对于如此大的有效负载大小，恶意行为者将很容易执行拒绝服务攻击(注:DOS攻击)，从而耗尽计算机上所有可用的随机存取内存(RAM)。保持合理最大有效负载大小会使内存耗尽攻击更难执行

清单4-7介绍了字符串类型，它与Binary一样，实现了`Payload`接口

```
type String string
// String实现的"Bytes"方法
func (m String) Bytes() []byte { return []byte(m) }
// String类型强制转换为string类型
func (m String) String() string { return string(m) }
// 和Binary的WriteTo方法类似
func (m String) WriteTo(w io.Writer) (int64, error) {
 // 第一个字节是StringType
 err := binary.Write(w, binary.BigEndian, StringType) // 1-byte type
 if err != nil {
 return 0, err
 }
 var n int64 = 1
 // 它将String强制转换为字节切片
 err = binary.Write(w, binary.BigEndian, uint32(len(m))) // 4-byte size
 if err != nil {
 return n, err
 }
 n += 4
 // 写入writer
 o, err := w.Write([]byte(m)) // payload
 return n + int64(o), err
}
```



```
The String implementation's Bytes method casts the String to a byte 
slice. The String method casts the String type to its base type, string. The 
String type's WriteTo method is like Binary's WriteTo method except the 
first byte written is the StringType and it casts the String to a byte slice 
before writing it to the writer.
Listing 4-8 finishes up the String type's Payload implementation
```



`String`实现的`Bytes`方法将字符串强制转换为字节切片，`String`方法将`String`类型强制转换为`string`类型(go语言是强类型，String别名必须通过string(xx)转换否则编译不过)，`String`类型的`WriteTo`方法与`Binary`的WriteTo方法类似，除了写入的第一个字节是` StringType`，它将字符串强制转换为字节切片

清单4-8完成了字符串类型的有效负载实现

```
func (m *String) ReadFrom(r io.Reader) (int64, error) {
	var typ uint8
	err := binary.Read(r, binary.BigEndian, &typ) // 1-byte type
	if err != nil {
		return 0, err
	}
	var n int64 = 1
	// 将typ变量与StringType进行比较
	if typ != StringType {
		return n, errors.New("invalid String")
	}
	var size uint32
	err = binary.Read(r, binary.BigEndian, &size) // 4-byte size
	if err != nil {
		return n, err
	}
	n += 4
	buf := make([]byte, size)
	o, err := r.Read(buf) // payload
	if err != nil {
		return n, err
	}
	//从r中Read的值强制转换为 String
	*m = String(buf)
	return n + int64(o), nil
}
```

```
Here, too, String's ReadFrom method is like Binary's ReadFrom method, with two exceptions. First, the method compares the typ variable against the StringType before proceeding. Second, the method casts the value read from the reader to a String.
```

​	在这里，`String`的`ReadFrom`方法也与`Binary`的ReadRrom方法一样，有两个例外。首先，该方法在继续之前将`typ`变量与`StringType `进行比较。其次，该方法将从读取器(`reader`)读取(`Read`)的值强制转换为`String`。

```
All that's left to implement is a way to read arbitrary data from a network connection and use it to constitute one of our two types. For that, we 
turn to Listing 4-9.
```

所要实现的只是一种从网络连接中读取任意数据并使用其构成我们两种类型的方法之一。为此，我们将转向清单4-9

```
//将读取器中的字节解码为二进制类型或字符串类型(
func decode(r io.Reader) (Payload, error) {
	var typ uint8
	// 获取类型typ
	err := binary.Read(r, binary.BigEndian, &typ)
	if err != nil {
		return nil, err
	}
	// 创建一个Payload变量
	var payload Payload
	//从读取器中读取的是否为预期的类型常量BinaryType or StringType
	switch typ {
	case BinaryType:
		payload = new(Binary)
	case StringType:
		payload = new(String)
	default:
		return nil, errors.New("unknown type")
	}
	_, err = payload.ReadFrom(
		// 已经读到的字节与读取器连接起来
		io.MultiReader(bytes.NewReader([]byte{typ}), r))
	if err != nil {
		return nil, err
	}
	return payload, nil
}

```

```
The decode function accepts an io.Reader and returns a Payload interface and an error interface. If decode cannot decode the bytes read from the reader into a Binary or String type, it will return an error along with a nil Payload.

```

这个解码函数接受一个`io.Reader`并返回一个`Payload`接口和一个`error`接口,如果解码无法将从读取器读取的字节解码为`Binary`或`String`类型，它将返回一个`error`以及一个零值**`Payload`**

```
You must first read a byte from the reader to determine the type and create a payload variable to hold the decoded type. If the type you read from the reader is an expected type constant , you assign the corresponding type to the payload variable.
```

您必须首先从阅读器读取一个字节，以确定类型并创建一个有效负载变量来保存已解码的类型，如果从阅读器读取的类型是预期的类型常量，则将相应的类型分配给负载变量

```
You now have enough information to finish decoding the binary data from the reader into the payload variable by using its ReadFrom method. But you have a problem here. You cannot simply pass the reader to the ReadFrom method. You've already read a byte from it corresponding to the type, yet the ReadFrom method expects the first byte it reads to be the type as well. Thankfully, the io package has a helpful function you can use: MultiReader. We cover io.MultiReader in more detail later in this chapter, but here you use it to concatenate the byte you've already read with the reader. From the ReadFrom method's perspective, it will read the bytes in the sequence it expects
```

现在您有了足够的信息，可以使用阅读器的`ReadFrom`方法将二进制数据解码为**payload**变量。但你这里有个问题。您不能简单地将读取器传递给ReadFrom方法。您已经从它中读取了对应于该类型的一个字节，然而ReadFrom方法期望它读取的第一个字节也是该类型。值得庆幸的是，io包有一个有用的功能可以使用:MultiReader ,我们覆盖的`io.MultiReader`将在本章后面详细介绍，但是这里使用它将已经读到的字节与读取器连接起来。从ReadFrom方法的角度来看，它将按照预期的顺序读取字节

```
Although the use of io.MultiReader shows you how to inject bytes back into a reader, it isn't optimal in this use case. The proper fix is to remove each type's need to read the first byte in its ReadFrom method. Then, the ReadFrom method would read only the 4-byte size and the payload, eliminating the need toinject the type byte back into the reader before passing it on to ReadFrom. As an exercise, I recommend you refactor the code to eliminate the need for io.MultiReader.
```



虽然`io.MultiReader`向您展示了如何将字节注入回读取器，在这个用例中它不是最佳的。正确的解决方法是删除每种类型在其`ReadFrom`方法中读取第一个字节的需要。然后，`ReadFrom`方法将只读取4字节大小和负载，而不需要在将类型字节传递给`ReadFrom`之前将其注入回读取器。作为练习，我建议您重构代码，以消除对`io.MultiReader`的需要。

```
Let's see the decode function in action in the form of a test. Listing 4-10 illustrates how you can send your two distinct types over a network connection and properly decode them back into their original type on the receiver's end.
```

让我们以测试的形式来看看解码函数的实际作用。清单4-10演示了如何通过网络连接发送两个不同的类型，并在接收端正确地将它们解码回原来的类型。

```go
package ch04

import (
	"bytes"
	"encoding/binary"
	"net"
	"reflect"
	"testing"
)

// 创建测试有效负载测试
func TestPayloads(t *testing.T) {
	// 创建两个BinaryType
	b1 := Binary("Clear is better than clever.")
	b2 := Binary("Don't panic.")
	// 创建一个String
	s1 := String("Errors are values.")
	//定义三个 Payload接口类型的指针
	payloads := []Payload{&b1, &s1, &b2}
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		for _, p := range payloads {
			// 将每一个payload写入writer中
			_, err = p.WriteTo(conn)
			if err != nil {
				t.Error(err)
				break
			}
		}
	}()
```

```
Your test should first create at least one of each type. You create two Binary types and one String type. Next, you create a slice of Payload interfaces and add pointers to the Binary and String types you created. You then create a listener that will accept a connection and write each type in the payloads slice to it.
This is a good start. Let's finish up the client side of the test in Listing 4-11
```

您的测试首先应该至少创建每种类型(BinaryType or StringType )中的一个。您创建了两个Binary类型和一个String类型。接下来，创建一个Payload接口片段，并添加指向所创建的Binary和String类型的指针。然后创建一个监听器，该监听器将接受连接，并将有效负载片中的每种类型写入该监听器。

这是一个良好的开端。让我们完成清单4-11中测试的客户端

```go

	// 完成测试payload测试
	// client发起一个listener的连接
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for i := 0; i < len(payloads); i++ {
		// 调用编码函数
		actual, err := decode(conn)
		if err != nil {
			t.Fatal(err)
		}
		// 将解码的类型与服务器发送的类型比较
		if expected := payloads[i]; !reflect.DeepEqual(expected, actual) {
			t.Errorf("value mismatch: %v != %v", expected, actual)
			continue
		}
		//使用-v标志运行测试，以查看类型及其值
        // go test -v types_test.go
		t.Logf("[%T] %[1]q", actual)
	}
}

```



```
You know how many types to expect in the payloads slice, so you initiate a connection to the listener and attempt to decode each one. Finally, your test compares the type you decoded with the type the server sent. If there's any discrepancy with the variable type or its contents, the test fails. You can run the test with the -v flag to see the type and its value.Let's make sure the Binary type enforces the maximum payload size in Listing 4-12.
```

您知道在有效负载片中期望有多少种类型，因此您发起一个到监听器的连接，并尝试对每个类型进行解码。最后，测试将解码的类型与服务器发送的类型进行比较。如果与变量类型或其内容有任何差异，则测试失败。您可以使用-v标志运行测试，以查看类型及其值。让我们确保Binary类型强制执行清单4-12中的最大负载大小。

```

// 测试最大payload大小
func TestMaxPayloadSize(t *testing.T) {
	// 创建一个 bytes.Buffer对象
	buf := new(bytes.Buffer)
	// 写入一个字节
	err := buf.WriteByte(BinaryType)
	if err != nil {
		t.Fatal(err)
	}
	// 1
	err = binary.Write(buf, binary.BigEndian, uint32(1<<30)) // 1 GB
	if err != nil {
		t.Fatal(err)
	}
	var b Binary
	_, err = b.ReadFrom(buf)
	// 2
	if err != ErrMaxPayloadSize {
		t.Fatalf("expected ErrMaxPayloadSize; actual: %v", err)
	}
}
```

```
This test starts with the creation of a bytes.Buffer containing the BinaryType byte and a 4-byte, unsigned integer indicating the payload is 1GB. When this buffer is passed to the Binary type's ReadFrom method, you receive the ErrMaxPayloadSize error in return. The test cases in Listings 4-10 and 4-11 should cover the use case of a payload that is less than the maximum size, but I encourage you to modify this test to make sure that's the case.

```

这个测试从创建一个`bytes.Buffer`开始。包含BinaryType字节和指示有效负载为 `1GB`的4字节无符号整数的缓冲区。当这个缓冲区被传递给Binary类型的ReadFrom方法时，返回ErrMaxPayloadSize错误。清单4-10和4-11中的测试用例应该涵盖小于最大大小的有效负载的用例，但我鼓励你修改这个测试以确保这是真的。

- 完整示例

```go
package ch04

import (
	"bytes"
	"encoding/binary"
	"net"
	"reflect"
	"testing"
)

// 创建测试有效负载测试
func TestPayloads(t *testing.T) {
	//  1
	b1 := Binary("Clear is better than clever.")
	b2 := Binary("Don't panic.")
	//2
	s1 := String("Errors are values.")
	//3
	payloads := []Payload{&b1, &s1, &b2}
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		for _, p := range payloads {
			//4
			_, err = p.WriteTo(conn)
			if err != nil {
				t.Error(err)
				break
			}
		}
	}()
	// 完成测试payload测试
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for i := 0; i < len(payloads); i++ {
		// 2
		actual, err := decode(conn)
		if err != nil {
			t.Fatal(err)
		}
		// 3
		if expected := payloads[i]; !reflect.DeepEqual(expected, actual) {
			t.Errorf("value mismatch: %v != %v", expected, actual)
			continue
		}
		//4
		t.Logf("[%T] %[1]q", actual)
	}
}

// 测试最大payload大小
func TestMaxPayloadSize(t *testing.T) {
	buf := new(bytes.Buffer)
	err := buf.WriteByte(BinaryType)
	if err != nil {
		t.Fatal(err)
	}
	// 1
	err = binary.Write(buf, binary.BigEndian, uint32(1<<30)) // 1 GB
	if err != nil {
		t.Fatal(err)
	}
	var b Binary
	_, err = b.ReadFrom(buf)
	// 2
	if err != ErrMaxPayloadSize {
		t.Fatalf("expected ErrMaxPayloadSize; actual: %v", err)
	}
}

```

**Handling Errors While Reading and Writing Data** 

- **处理读写数据时的错误**

```
Unlike writing to file objects, writing to network connections can be unreliable, especially if your network connection is spotty. Files don't often return errors while you're writing to them, but the receiver on the other end of a network connection may abruptly disconnect before you write your entire payload.
```

与对文件对象的写入不同，对网络连接的写入可能是不可靠的，特别是在网络连接不稳定的情况下。当您向文件写入时，文件通常不会返回错误，但是在您写入整个有效负载之前，网络连接另一端的接收者可能会突然断开连接。

```
Not all errors returned when reading from or writing to a network connection are permanent. The connection can recover from some errors. For example, writing data to a network connection where adverse network conditions delay the receiver's ACK packets, and where your connection times out while waiting to receive them, can result in a temporary error. This can occur if someone temporarily unplugs a network cable between you and the receiver. In that case, the network connection is still active, and you can either attempt to recover from the error or gracefully terminate your end of the connection.

```

​	并非所有从网络连接读取或写入时返回的错误都是永久性的。连接可以从某些错误中恢复。例如，向网络连接写入数据时，不良的(较差的)网络条件会延迟接收方的ACK数据包，而在等待接收这些数据包时，您的连接会超时，这可能会导致临时错误。如果有人临时拔掉你和接收器之间的网线，就会发生这种情况。在这种情况下，网络连接仍然是活动的，您可以尝试从错误中恢复，或者优雅地终止连接。

```
Listing 4-13 illustrates how to check for temporary errors while writing 
data to a network connection
```





```
Since you might receive a transient error when writing to a network 
connection, you might need to retry a write operation. One way to account 
for this is to encapsulate the code in a for loop. This makes it easy to 
retry the write operation, if necessary.
```

由于在写入网络连接时可能会收到临时错误，因此可能需要重试写操作。解释这一点的一种方法是将代码封装在一个循环中。如有必要，可以轻松地重试写操作

```
To write to the connection, you pass a byte slice to the connection’s Write method as you would to any other io.Writer. This returns the number of bytes written and an error interface. If the error interface is not nil, you check whether the error implements the net.Error interface by using a type assertionand check whether the error is temporary. If the net.Error’s Temporary method returns true, the code makes another write attempt by iterating around the for loop. If the error is permanent, the code returns the error. A successful write breaks out of the loop.
```

要写入连接，将字节切片传递给连接的写入方法，就像向任何其他`io.Writer`一样。这将返回已写入的字节数和一个`error`接口。如果错误接口不是`nil`，则您通过使用类型断言来检查该`error`是否实现了` net.Error`接口，并检查错误是否为临时性。如果`net.Error`的`Temporary`方法返回true，该代码通过迭代for循环进行另一次写尝试。如果错误是永久性的，则代码返回错误。成功的写操作会中断循环。



**Creating Robust Network Applications by Using the** **io Package**

- 通过使用io包创建鲁棒性的网络应用程序

