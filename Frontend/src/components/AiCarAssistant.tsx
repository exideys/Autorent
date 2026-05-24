import { FormEvent, useState } from "react";
import { Bot, CarFront, Loader2, Send, Sparkles } from "lucide-react";

import { Car, recommendCar } from "../api/ai";

export function AiCarAssistant() {
  const [message, setMessage] = useState("");
  const [answer, setAnswer] = useState("");
  const [cars, setCars] = useState<Car[]>([]);
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const trimmedMessage = message.trim();

    if (!trimmedMessage) {
      setError("Please describe what kind of car you need.");
      return;
    }

    setIsLoading(true);
    setError("");
    setAnswer("");
    setCars([]);

    try {
      const result = await recommendCar(trimmedMessage);

      setAnswer(result.answer);
      setCars(result.cars ?? []);
    } catch (err) {
      const errorMessage =
        err instanceof Error
          ? err.message
          : "AI assistant is temporarily unavailable. Please try again later.";

      setError(errorMessage);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <section className="relative overflow-hidden rounded-[2rem] border border-white/10 bg-slate-950/80 p-6 shadow-2xl shadow-cyan-950/30 backdrop-blur md:p-8">
      <div className="absolute right-0 top-0 h-40 w-40 rounded-full bg-cyan-500/20 blur-3xl" />
      <div className="absolute bottom-0 left-0 h-40 w-40 rounded-full bg-violet-500/20 blur-3xl" />

      <div className="relative z-10 grid gap-8 lg:grid-cols-[0.9fr_1.1fr]">
        <div className="space-y-5">
          <div className="inline-flex items-center gap-2 rounded-full border border-cyan-400/30 bg-cyan-400/10 px-4 py-2 text-sm font-medium text-cyan-100">
            <Sparkles className="h-4 w-4" />
            AI car assistant
          </div>

          <div className="space-y-3">
            <h2 className="text-3xl font-bold tracking-tight text-white md:text-4xl">
              Find the right car with AI
            </h2>

            <p className="max-w-xl text-sm leading-6 text-slate-300 md:text-base">
              Describe your trip, budget, number of passengers, preferred class,
              fuel type, or car body. The assistant will check available cars and
              recommend suitable options.
            </p>
          </div>

          <div className="rounded-3xl border border-white/10 bg-white/[0.03] p-4 text-sm text-slate-300">
            <p className="mb-2 font-medium text-white">Example requests:</p>
            <ul className="space-y-1">
              <li>“для сімʼї з 5 людей до 240 доларів на день”</li>
              <li>“premium car for business under 230 dollars per day”</li>
              <li>“електрокар на 5 місць”</li>
            </ul>
          </div>
        </div>

        <div className="rounded-[1.5rem] border border-white/10 bg-white/[0.04] p-4 shadow-xl">
          <form onSubmit={handleSubmit} className="space-y-4">
            <label
              htmlFor="ai-car-request"
              className="block text-sm font-medium text-slate-200"
            >
              Your request
            </label>

            <textarea
              id="ai-car-request"
              value={message}
              onChange={(event) => setMessage(event.target.value)}
              placeholder="For a family of 5 people up to $240 per day"
              className="min-h-28 w-full resize-none rounded-2xl border border-white/10 bg-slate-900/80 px-4 py-3 text-sm text-white outline-none transition placeholder:text-slate-500 focus:border-cyan-400 focus:ring-2 focus:ring-cyan-400/20"
            />

            <button
              type="submit"
              disabled={isLoading}
              className="inline-flex w-full items-center justify-center gap-2 rounded-2xl bg-cyan-400 px-5 py-3 text-sm font-semibold text-slate-950 transition hover:bg-cyan-300 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {isLoading ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Searching cars...
                </>
              ) : (
                <>
                  <Send className="h-4 w-4" />
                  Ask AI
                </>
              )}
            </button>
          </form>

          {error && (
            <div className="mt-5 rounded-2xl border border-red-400/30 bg-red-500/10 p-4 text-sm text-red-100">
              {error}
            </div>
          )}

          {answer && (
            <div className="mt-5 rounded-2xl border border-cyan-400/20 bg-cyan-400/10 p-4">
              <div className="mb-2 flex items-center gap-2 text-sm font-semibold text-cyan-100">
                <Bot className="h-4 w-4" />
                AI recommendation
              </div>

              <p className="whitespace-pre-line text-sm leading-6 text-slate-100">
                {answer}
              </p>
            </div>
          )}

          {cars.length > 0 && (
            <div className="mt-5 grid gap-3">
              {cars.map((car) => (
                <article
                  key={car.id}
                  className="rounded-2xl border border-white/10 bg-slate-900/80 p-4 transition hover:border-cyan-400/40 hover:bg-slate-900"
                >
                  <div className="flex items-start justify-between gap-4">
                    <div>
                      <div className="flex items-center gap-2 text-white">
                        <CarFront className="h-4 w-4 text-cyan-300" />
                        <h3 className="font-semibold">
                          {car.brand} {car.model}
                        </h3>
                      </div>

                      <p className="mt-1 text-sm text-slate-400">
                        {car.car_class} · {car.body_type} · {car.fuel_type}
                      </p>
                    </div>

                    <div className="text-right">
                      <p className="text-lg font-bold text-cyan-200">
                        ${car.price_per_day}
                      </p>
                      <p className="text-xs text-slate-500">per day</p>
                    </div>
                  </div>

                  <div className="mt-4 grid grid-cols-2 gap-2 text-xs text-slate-300 sm:grid-cols-4">
                    <div className="rounded-xl bg-white/[0.04] px-3 py-2">
                      <span className="block text-slate-500">Seats</span>
                      {car.seats}
                    </div>

                    <div className="rounded-xl bg-white/[0.04] px-3 py-2">
                      <span className="block text-slate-500">Year</span>
                      {car.year}
                    </div>

                    <div className="rounded-xl bg-white/[0.04] px-3 py-2">
                      <span className="block text-slate-500">Power</span>
                      {car.horsepower} hp
                    </div>

                    <div className="rounded-xl bg-white/[0.04] px-3 py-2">
                      <span className="block text-slate-500">Deposit</span>$
                      {car.deposit}
                    </div>
                  </div>
                </article>
              ))}
            </div>
          )}
        </div>
      </div>
    </section>
  );
}