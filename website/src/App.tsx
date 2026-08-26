import Header from "./components/Header";

function App() {
  return (
    <>
      <div className="bg-[#0A0A0A] h-dvh">
        <Header />
        <div className="flex pt-4 justify-center items-center">
          <p className="text-4xl text-[#00FF41] place-content-center">
            Welcome to Wick the compiler!
          </p>
        </div>
      </div>
    </>
  );
}

export default App;
