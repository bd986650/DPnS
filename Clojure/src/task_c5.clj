(ns task-c5)

(defn create-forks [n]
  (mapv #(ref {:id % :uses 0 :owner nil})
        (range n)))

(def tx-attempts (atom 0))

(defn try-acquire-forks [id left right]
  (dosync
    (swap! tx-attempts inc)
    (when (and (nil? (:owner @left))
               (nil? (:owner @right)))
      (doseq [fork [left right]]
        (alter fork #(-> %
                         (update :uses inc)
                         (assoc :owner id))))
      true)))

(defn release-forks [left right]
  (dosync
    (doseq [fork [left right]]
      (alter fork assoc :owner nil))))

(defn dine-once [id left right think-ms eat-ms]
  (Thread/sleep think-ms)
  (loop []
    (if (try-acquire-forks id left right)
      (do
        (Thread/sleep eat-ms)
        (release-forks left right))
      (recur))))

(defn spawn-philosopher
  [id left right think-ms eat-ms meals]
  (future
    (dotimes [_ meals]
      (dine-once id left right think-ms eat-ms))))

(defn run-simulation
  [philosophers think-ms eat-ms meals]
  (reset! tx-attempts 0)
  (let [forks (create-forks philosophers)
        start (System/nanoTime)
        workers (mapv
                  (fn [id]
                    (spawn-philosopher
                      id
                      (nth forks id)
                      (nth forks (mod (inc id) philosophers))
                      think-ms eat-ms meals))
                  (range philosophers))]
    (doseq [w workers] @w)

    (let [elapsed (/ (- (System/nanoTime) start) 1e9)
          attempts @tx-attempts
          total-meals (* philosophers meals)]
      {:philosophers philosophers
       :meals-per-phil meals
       :think-ms think-ms
       :eat-ms eat-ms
       :elapsed-sec elapsed
       :tx-attempts attempts
       :extra-attempts (- attempts total-meals)
       :forks (mapv deref forks)})))


(defn print-report [result]
  (println "Philosophers:" (:philosophers result))
  (println "Meals per philosopher:" (:meals-per-phil result))
  (println "Think:" (:think-ms result) "ms | Eat:" (:eat-ms result) "ms")
  (println "Elapsed:" (:elapsed-sec result) "sec")
  (println "Transaction attempts:" (:tx-attempts result))
  (println "Extra attempts:" (:extra-attempts result))
  (doseq [f (:forks result)]
    (println "Fork" (:id f) "uses:" (:uses f))))
